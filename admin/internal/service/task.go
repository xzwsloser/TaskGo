package service

import (
	"fmt"

	"github.com/xzwsloser/TaskGo/admin/internal/model/request"
	"github.com/xzwsloser/TaskGo/model"
	"github.com/xzwsloser/TaskGo/pkg/etcdclient"
	"github.com/xzwsloser/TaskGo/pkg/logger"
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

func (*TaskService) SearchTaskLog(r *request.ReqTaskLogSearch) ([]model.TaskLog, int64, error) {
	taskLog := &model.TaskLog{}
	taskLog.NodeUUID = r.NodeUUID
	taskLog.Name = r.Name
	taskLog.TaskId = r.TaskId
	taskLog.Success = r.Success

	pageSize, page := r.PageSize, r.Page
	return taskLog.FindAndPage(page, pageSize)
}

func (*TaskService) ExecOnce(once *request.ReqTaskOnce) error {
	_, err := etcdclient.GetEtcdClient().
						 PutWithTTL(fmt.Sprintf(etcdclient.KeyEtcdOnceFormat, once.TaskId),
									   once.NodeUUID, 60)
	return err
}

func (*TaskService) GetNotAssignTasks() ([]model.Task, error) {
	t := &model.Task{}
	return t.GetNotAssignedTasks()
}





