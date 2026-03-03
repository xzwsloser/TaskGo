package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"syscall"
	"time"
	"github.com/jakecoffman/cron"
	"github.com/xzwsloser/TaskGo/model"
	"github.com/xzwsloser/TaskGo/node/server/handler"
	"github.com/xzwsloser/TaskGo/pkg/config"
	"github.com/xzwsloser/TaskGo/pkg/etcdclient"
	"github.com/xzwsloser/TaskGo/pkg/logger"
	"github.com/xzwsloser/TaskGo/pkg/utils"
)

const (
	DefaultTaskSize	   = 8
)

var (
	ErrNodeExists		= errors.New("Node Already Register In Etcd")
	ErrNodeInfoMarshal	= errors.New("Failed To Marshal Node Info")
)

// @Description: Node Server To Exec Task
type NodeServer struct {
	*etcdclient.Register	// Service Register
	*model.Node				// Node Info
	*cron.Cron				// Cron Task

	tasks	handler.Tasks  // Tasks To Exec
}

// @Description: Create Node Server To Exec Task
func NewNodeServer() (*NodeServer, error) {
	nodeUUID, err := utils.GenerateUUID()
	if err != nil {
		return nil, err
	}

	ip, err := utils.GetLocalIP()
	if err != nil {
		return nil, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = nodeUUID
		err = nil
	}

	return &NodeServer{
		Node: &model.Node{
			UUID: nodeUUID,
			PID: strconv.Itoa(os.Getpid()),
			IP: ip.String(),
			Hostname: hostname,
			UpTime: time.Now().Unix(),
			Status: model.NodeConnSuccess,
			Version: config.GetConfig().System.Version,
		},
		Cron: cron.New(),
		tasks: make(handler.Tasks, DefaultTaskSize),
		Register: etcdclient.NewRegister(config.GetConfig().System.NodeTtl),
	}, nil
}


// @Description: Check If The Node Exists
// If Exists, Return PID, Else Return -1
func (srv *NodeServer) exist(nodeUUID string) (pid int, err error) {
	resp, err := etcdclient.GetEtcdClient().Get(
		fmt.Sprintf(etcdclient.KeyEtcdNodeFormat, nodeUUID))
	if err != nil {
		return 
	}

	if len(resp.Kvs) == 0 {
		return -1, nil
	}

	if pid, err = strconv.Atoi(string(resp.Kvs[0].Value)); err != nil {
		if _, err = etcdclient.GetEtcdClient().Delete(fmt.Sprintf(etcdclient.KeyEtcdNodeFormat, nodeUUID)) ; err != nil {
			return 
		}

		return -1, nil
	}

	// Confirm The Process Is Exists
	p, err := os.FindProcess(pid)
	if err != nil {
		return -1, nil
	}

	if p != nil && p.Signal(syscall.Signal(0)) == nil {
		// Find The Process
		return 
	}

	return -1, nil
}

// @Description: Service Register Using Etcd
func (srv *NodeServer) RegisterNodeServer() error {
	pid, err := srv.exist(srv.UUID)
	if err != nil {
		return err
	}

	if pid != -1 {
		err = ErrNodeExists
		logger.GetLogger().Error(err.Error())
		return err
	}

	buf, err := json.Marshal(&srv.Node)
	if err != nil {
		err = ErrNodeInfoMarshal
		logger.GetLogger().Error(err.Error())
		return err
	}

	// Register Node
	err = srv.Register.RegisterService(fmt.Sprintf(etcdclient.KeyEtcdNodeFormat, srv.UUID), string(buf))
	if err != nil {
		logger.GetLogger().Error(err.Error())
		return err
	}

	return nil
}

// @Description: Update The Status Of Task On Node Server
func (srv *NodeServer) Down() {
	srv.Status = model.NodeConnFail
	srv.DownTime = time.Now().Unix()
	err := srv.Node.Update()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("Failed To Update Node Info: %v\n", err))
	}

	taskModel := &model.Task{
		Status: model.TaskStatusNotAssigned,
	}
	err = taskModel.UpdateByNodeUUID(srv.UUID)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("Failed To Update Task Info: %v\n", err))
	}
}

// @Description: Stop Node Server, Delete Task
func (srv *NodeServer) Stop(any) {
	srv.Down()
	_, err := etcdclient.GetEtcdClient().Delete(fmt.Sprintf(etcdclient.KeyEtcdNodeFormat, srv.UUID))
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("Node: %s Failed To Delete Etcd Key: %v", srv.UUID, err))
	}

	_, err = etcdclient.GetEtcdClient().Delete(fmt.Sprintf(etcdclient.KeyEtcdSystemGet, srv.UUID))
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("Node: %s Failed To Delete System Get Info: %v", srv.UUID, err))
	}

	etcdclient.GetEtcdClient().Close()
	srv.Cron.Stop()
}

func (srv *NodeServer) taskCronName(taskID int) string {
	return fmt.Sprintf(srv.UUID + "/%d", taskID)
}

// @Description: Add Task Into Node Exec List
func (srv *NodeServer) addTask(task *handler.Task) {
	if err := task.Check() ; err != nil {
		logger.GetLogger().Error(fmt.Sprintf("Task [%d] Check Error: %v", task.ID, err))
		return
	}

	// Run Preset Script
	// Environment Problem
	if task.Type == model.TaskTypeCmd {
		for _, id := range task.ScriptIDArray {
			script := &model.Script{ID: id}
			err := script.FindById()
			if err != nil {
				logger.GetLogger().Error(fmt.Sprintf("Task [%d] Not Find: %v", task.ID, err))
				continue
			}
			err = script.Check()
			if err != nil {
				logger.GetLogger().Error(fmt.Sprintf("Script [%d] Check Error: %v", id, err))
				continue
			}

			result, err := handler.RunPresetScript(script)
			if err != nil {
				logger.GetLogger().Warn(fmt.Sprintf("Task [%d] Run Script [%d] Error: %v", task.ID, id, err))
			}
			logger.GetLogger().Info(fmt.Sprintf("Task [%d] Run Script [%d] Result: %s", task.ID, id, result))
		}
	}

	taskFunc := handler.CreateTaskCallback(task)
	if taskFunc == nil {
		logger.GetLogger().Error(fmt.Sprintf("Task [%d] Failed To Create Callback Func", task.ID))
		return
	}

	err := utils.PanicToError(func() {
		srv.Cron.AddFunc(task.Spec, taskFunc, srv.taskCronName(task.ID))
	})

	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("Failed To Add Func To Cron: %v", err))
	}
}

// @Description: Delete Task From Node Exec List
func (srv *NodeServer) deleteTask(taskID int) {
	if _, ok := srv.tasks[taskID] ; ok {
		srv.Cron.RemoveJob(srv.taskCronName(taskID))
		delete(srv.tasks, taskID)
		return
	}
}

// @Description: Modify Task
func (srv *NodeServer) modifyTask(task *handler.Task) {
	prevTask, ok := srv.tasks[task.ID]
	if !ok {
		srv.addTask(task)
		return
	}
	srv.deleteTask(prevTask.ID)
	srv.tasks[task.ID] = task
	srv.addTask(task)
}

// @Description: Load Task From Etcd
func (srv *NodeServer) loadTasks() (err error) {
	defer func() {
		if r := recover() ; r != nil {
			logger.GetLogger().Error(fmt.Sprintf("Node Server Load Task Failed: %v", err))
		}
	}()

	tasks, err := handler.GetTasks(srv.UUID)
	if err != nil || len(tasks) == 0 {
		return 
	}

	srv.tasks = tasks
	for _, task := range tasks {
		task.AssociatedNodeInfo(model.TaskStatusAssigned, srv.UUID, srv.Hostname, srv.IP)
		srv.addTask(task)
	}

	return 
}

// TODO: Node Server Listen And Run
func (srv *NodeServer) Run() (err error) {
	return 
}

