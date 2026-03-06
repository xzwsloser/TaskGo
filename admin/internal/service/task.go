package service

import (
	"fmt"
	"time"

	"github.com/xzwsloser/TaskGo/admin/internal/model/request"
	"github.com/xzwsloser/TaskGo/admin/internal/model/resp"
	"github.com/xzwsloser/TaskGo/model"
	"github.com/xzwsloser/TaskGo/pkg/dbclient"
	"github.com/xzwsloser/TaskGo/pkg/etcdclient"
	"github.com/xzwsloser/TaskGo/pkg/logger"
	"github.com/xzwsloser/TaskGo/pkg/utils"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type TaskService struct {
}

const (
	MaxTaskCount	= 10000
)

// @Description: Auto Allocate Node To Exec Task
// Select The Node Has The Lease Task
func (*TaskService) AutoAllocateNode() string {
	nodeList := nodeWatcherService.NodeListToArray()
	allocatedUUID := ""
	minTaskCount := MaxTaskCount
	for _, nodeUUID := range nodeList {
		node := &model.Node{UUID: nodeUUID}
		err := node.FindByUUID()
		if err != nil {
			continue
		}

		if node.Status == model.NodeConnFail {
			delete(nodeWatcherService.nodeLists, nodeUUID)
			continue
		}

		count, err := nodeWatcherService.GetTasksCount(nodeUUID)
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("node [%s] get task count failed: %v", nodeUUID, err))
			continue
		}

		if minTaskCount > count {
			minTaskCount = count
			allocatedUUID = nodeUUID
		}
	}

	return allocatedUUID
}

// @Description: Search For Task By Page And Condition
func (*TaskService) Search(r *request.ReqTaskSearch) ([]model.Task, int64, error) {
	task := &model.Task{}
	task.ID = r.ID
	task.RunOn = r.RunOn
	task.Name = r.Name
	task.Type = r.Type
	task.Status = r.Status
	page, pageSize := r.Page, r.PageSize

	return task.FindAndPage(page, pageSize)
}

// @Description: Search For Task Log By Page And Condition
func (*TaskService) SearchTaskLog(r *request.ReqTaskLogSearch) ([]model.TaskLog, int64, error) {
	taskLog := &model.TaskLog{}
	taskLog.NodeUUID = r.NodeUUID
	taskLog.Name = r.Name
	taskLog.TaskId = r.TaskId
	taskLog.Success = r.Success

	pageSize, page := r.PageSize, r.Page
	return taskLog.FindAndPage(page, pageSize)
}

// @Description: Exec For The Certain Task Immediately
func (*TaskService) ExecOnce(once *request.ReqTaskOnce) error {
	_, err := etcdclient.GetEtcdClient().
						 PutWithTTL(fmt.Sprintf(etcdclient.KeyEtcdOnceFormat, once.TaskId),
									   once.NodeUUID, 60)
	return err
}

// @Description: Get Not Assign Task
func (*TaskService) GetNotAssignTasks() ([]model.Task, error) {
	t := &model.Task{}
	return t.GetNotAssignedTasks()
}

// @Description: Get Task Count Of Today
func (*TaskService) GetTodayTaskExecCount(success int) (int64, error) {
	startTime := utils.GetTodayTimeStamp()
	task := &model.TaskLog{}
	count, err := task.GetTaskCountAfter(startTime, success)
	return count, err
}

// @Description: Get Task At Certain Period Of Time
func (*TaskService) GetTaskExecCount(start, end int64, success int) ([]resp.RspDateCount, error) {
	task := &model.TaskLog{}
	mdc, err := task.GetTaskExecCount(start, end, success)
	if err != nil {
		return nil, err
	}
	dc := make([]resp.RspDateCount, len(mdc))
	for idx := range mdc {
		dc[idx] = resp.RspDateCount{
			Date: mdc[idx].Date,
			Count: mdc[idx].Count,
		}
	}
	return dc, nil
}

// @Description: Get The Number Of Running Task
func (*TaskService) GetRunningTaskCount() (int64, error) {
	resp, err := etcdclient.GetEtcdClient().Get(etcdclient.KeyEtcdProcPrefix, clientv3.WithPrefix(), clientv3.WithCountOnly())
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("failed to count running task: %s", err.Error()))
		return 0, err
	}
	return resp.Count, nil
}

func cleanupLogs(expirationTime int64) error {
	sql := fmt.Sprintf("delete from %s where start_time < ?", model.TaskGoTaskLogTableName)
	return dbclient.GetMysqlDB().Exec(sql, time.Now().Unix()-expirationTime).Error
}

// @Description: Clear Log At Certain Interval
func RunLogCleaner(cleanPeriod time.Duration, 
	expiration int64) (close chan struct{}) {
	t := time.NewTicker(cleanPeriod)
	close = make(chan struct{})
	go func() {
		for {
			select {
			case <-t.C:
				err := cleanupLogs(expiration)
				if err != nil {
					logger.GetLogger().Error(fmt.Sprintf("clean up logs at time:%v error:%s", time.Now(), err.Error()))
				}
			case <-close:
				t.Stop()
				return
			}
		}
	}()
	return
}









