package handler

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/xzwsloser/TaskGo/model"
	"github.com/xzwsloser/TaskGo/pkg/etcdclient"
	"github.com/xzwsloser/TaskGo/pkg/logger"
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



