package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/xzwsloser/TaskGo/model"
	"github.com/xzwsloser/TaskGo/pkg/etcdclient"
	"github.com/xzwsloser/TaskGo/pkg/logger"
	"github.com/xzwsloser/TaskGo/pkg/notify"
	"github.com/xzwsloser/TaskGo/pkg/utils"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Task struct {
	*model.Task
}

type Tasks map[int]*Task

var (
	ErrNotFound = errors.New("Failed To Get Task In Etcd")
)

// @Description: Get The Task ID Registered In Etcd
func TaskKey(nodeUUID string, taskID int) string {
	return fmt.Sprintf(etcdclient.KeyEtcdTaskFormat, nodeUUID, taskID)
}

// @Description: Get Task Info And Version Split Cmd
func GetTaskAndRev(nodeUUID string, taskID int) (*Task, int64, error) {
	resp, err := etcdclient.GetEtcdClient().Get(TaskKey(nodeUUID, taskID))
	if err != nil {
		return nil, 0, err
	}

	if resp.Count == 0 {
		return nil, 0, ErrNotFound
	}

	task := &Task{}
	rev := resp.Kvs[0].ModRevision
	if err = json.Unmarshal(resp.Kvs[0].Value, &task) ; err != nil {
		return  task, rev, err
	}

	task.SplitCmd()
	return task, rev, err
}

// @Description: Get All Task From A Node
func GetTasks(nodeUUID string) (tasks Tasks, err error) {
	resp, err := etcdclient.GetEtcdClient().Get(fmt.Sprintf(etcdclient.KeyEtcdTaskPrefix, nodeUUID), clientv3.WithPrefix())
	if err != nil {
		return 
	}

	count := len(resp.Kvs)
	if count == 0 {
		return 
	}

	tasks = make(Tasks, count)

	for _, v := range resp.Kvs {
		task := &Task{}
		if err = json.Unmarshal(v.Value, task) ; err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("task [%s] Unmarshal error: %s\n", v.Key, err.Error()))
			continue
		}

		if err := task.Check() ; err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("task [%s] is invalid: %s\n", v.Key, err.Error()))
			continue 
		}

		tasks[task.ID] = task
	}

	return 
}

func WatchTasks(nodeUUID string) clientv3.WatchChan {
	return etcdclient.GetEtcdClient().Watch(
		fmt.Sprintf(etcdclient.KeyEtcdTaskPrefix, nodeUUID), clientv3.WithPrefix())
}

func GetTaskFromKey(key string) int {
	idx := strings.LastIndex(key, "/")
	if idx < 0 {
		return 0
	}

	taskID, err := strconv.Atoi(key[idx+1:])
	if err != nil {
		return 0
	}

	return taskID
}

// @Description: Insert Task Log (Task Info) Into DB
func (task *Task) CreateTaskLog() (int, error) {
	start := time.Now()
	taskLog := &model.TaskLog{
		Name: task.Name,
		TaskId: task.ID,
		Command: task.Command,
		IP: task.Ip,
		Hostname: task.Hostname,
		NodeUUID: task.RunOn,
		Spec: task.Spec,
		StartTime: start.Unix(),
	}

	return taskLog.Insert()
}

// @Description: Update After The Task Finished
func (task *Task) UpdateTaskLog(taskLogId int, 
								start time.Time, 
								output string, 
								retry int, 
								success bool) error {
	end := time.Now()
	taskLog := &model.TaskLog{
		ID: taskLogId,
		StartTime: start.Unix(),
		RetryTimes: retry,
		Success: success,
		Output: output,
		EndTime: end.Unix(),
	}

	return taskLog.Update()
}

func (task *Task) Success(taskLogId int, 
						  start time.Time, 
						  output string, 
						  retry int) error {
	return task.UpdateTaskLog(taskLogId, start, output, retry, true)
}

func (task *Task) Fail(taskLogId int, 
					   start time.Time, 
					   output string, 
					   retry int) error {
	return task.UpdateTaskLog(taskLogId, start, output, retry, false)
}

// @Description: Run Task And Recover
func (task *Task) RunWithRecover() {
	// recover
	defer func() {
		if r := recover() ; r != nil {
			const stackSize = 64 << 10 // 64 KB
			buf := make([]byte, stackSize)
			buf = buf[:runtime.Stack(buf, false)]
			logger.GetLogger().Warn(fmt.Sprintf("panic running task: %v\n%s", r, buf))
		}
	}()

	t := time.Now()
	taskLogId, err := task.CreateTaskLog()
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("Failed to write task log: %v\n", err))
	}
	h := NewHandler(task)
	if h == nil {
		logger.GetLogger().Error(fmt.Sprintf("Task: %d Has No Type!", task.ID))
		return 
	}

	result, runErr := h.Run(task)
	if runErr != nil {
		// Task Failed -> Notify User
		err = task.Fail(taskLogId, t, runErr.Error(), 0)
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("Failed to write task log: %v\n", err))
		}

		node := &model.Node{UUID: task.RunOn}
		err = node.FindByUUID()
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("Failed to find node: %v\n", err))
		}

		var notifiedUsers []string
		for _, userId := range task.NotifyToArray {
			userModel := &model.User{
				ID: userId,
			}
			err = userModel.FindById()
			if err != nil {
				logger.GetLogger().Warn(fmt.Sprintf("Failed to Find User: %d\n", userId))
				continue
			}

			notifiedUsers = append(notifiedUsers, userModel.Email)
		}

		msg := &notify.Message{
			Type: task.NotifyType,
			IP: fmt.Sprintf("%s:%s", node.IP, node.PID),
			Subject: fmt.Sprintf("任务 [%s] 执行失败", task.Name),
			Body: fmt.Sprintf("Task [%d] Run On Node [%s] Once Execute Failed, Output: %s, Error: %s", task.ID, task.RunOn, result, runErr.Error()),
			To: notifiedUsers,
			OccurTime: time.Now().Format(utils.TimeFormatSecond),
		}

		go notify.Send(msg)
	} else {
		// Task Success -> Record Log
		err = task.Success(taskLogId, t, result, 0)
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("Failed To Write Task Log: %v\n", err))
		}
	}
}

// @Description: Create Task Cron Callback Func
func CreateTaskCallback(task *Task) cron.FuncJob {
	h := NewHandler(task)
	if h == nil {
		logger.GetLogger().Error("Failed To Create Task Handler")
		return nil
	}

	taskFunc := func() {
		logger.GetLogger().Info(fmt.Sprintf("Start The Task [%s] Command [%s]", task.Name, task.Command))
		var execTimes int = 1
		if task.RetryTimes > 0 {
			execTimes += task.RetryTimes
		}

		var tryTimes int = 0
		var output string 
		var runErr error // Task Run Error
		var err error    // Func Run Error
		var taskLogId int
		t := time.Now()
		taskLogId, err = task.CreateTaskLog()
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("Failed To Write Task Log: %v\n", err))
		}

		for tryTimes < execTimes {
			output, runErr = h.Run(task)
			if runErr == nil {
				// Task Success
				err = task.Success(taskLogId, t, output, tryTimes)
				if err != nil {
					logger.GetLogger().Warn(fmt.Sprintf("Failed To Write Task Log: %v", err))
				}
				return 
			}

			tryTimes ++
			if tryTimes < execTimes {
				logger.GetLogger().Warn(fmt.Sprintf("Task Execution Failed: %v", runErr))
				if task.RetryInterval > 0 {
					time.Sleep(time.Duration(task.RetryInterval) * time.Second)
				} else {
					time.Sleep(time.Duration(tryTimes) * time.Minute)
				}
			}
		}

		// Task Failed
		err = task.Fail(taskLogId, t, runErr.Error(), execTimes-1)
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("Failed To Write Task Log: %v", err))
		}
		node := &model.Node{UUID: task.RunOn}
		err = node.FindByUUID()
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("Failed to find node: %v\n", err))
		}

		var notifiedUsers []string
		for _, userId := range task.NotifyToArray {
			userModel := &model.User{
				ID: userId,
			}
			err = userModel.FindById()
			if err != nil {
				logger.GetLogger().Warn(fmt.Sprintf("Failed to Find User: %d\n", userId))
				continue
			}

			notifiedUsers = append(notifiedUsers, userModel.Email)
		}

		msg := &notify.Message{
			Type: task.NotifyType,
			IP: fmt.Sprintf("%s:%s", node.IP, node.PID),
			Subject: fmt.Sprintf("任务 [%s] 执行失败", task.Name),
			Body: fmt.Sprintf("Task [%d] Run On Node [%s] Once Execute Failed, Output: %s, Error: %s", task.ID, task.RunOn, output, runErr.Error()),
			To: notifiedUsers,
			OccurTime: time.Now().Format(utils.TimeFormatSecond),
		}

		go notify.Send(msg)

	}

	return taskFunc
}



