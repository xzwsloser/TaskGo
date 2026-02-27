package handler

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/xzwsloser/TaskGo/model"
	"github.com/xzwsloser/TaskGo/pkg/config"
	"github.com/xzwsloser/TaskGo/pkg/etcdclient"
	"github.com/xzwsloser/TaskGo/pkg/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// @Description: Info Aboout Current Task In Execution
// Key: /taskGo/proc/<node_uuid>/<task_id>/<pid>
// Value: The Start Execution Time Of Task
//
// The key will expire automatically
// To Prevent the key from being cleared
// after the process exits unexpectedly
// the expiration time can be configured
type TaskProc struct {
	*model.TaskProc
}

func GetProcFromKey(key string) (proc *TaskProc, err error) {
	idList := strings.Split(key, "/")
	idListLen := len(idList)
	if idListLen < 5 {
		err = fmt.Errorf("Invalid Proc Key [%s]", key)
		return 
	}

	pid, err := strconv.Atoi(idList[idListLen-1])
	if err != nil {
		return 
	}

	taskId, err := strconv.Atoi(idList[idListLen-2])
	if err != nil {
		return 
	}

	proc = &TaskProc{
		TaskProc: &model.TaskProc{
			ID: pid,
			TaskID: taskId,
			NodeUUID: idList[idListLen-3],
		},
	}

	return 
}

func (tp *TaskProc) Key() string {
	return fmt.Sprintf(etcdclient.KeyEtcdProcFormat, tp.NodeUUID, tp.TaskID, tp.ID)
}

func (tp *TaskProc) delProc() error {
	_, err := etcdclient.GetEtcdClient().Delete(tp.Key())
	return err
}

func (tp *TaskProc) Stop() {
	if tp == nil {
		return 
	}

	// If Task Task Running Return 
	// Else Set tp.Running = 0
	if !atomic.CompareAndSwapInt32(&tp.Running, 1, 0) {
		return 
	}

	// Wait for other task execute
	tp.Wg.Wait()

	if err := tp.delProc() ; err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("proc del [%s] failed: %s\n", tp.Key(), err.Error()))
	}

}

func (tp *TaskProc) Start() error {
	// TaskProc Started
	if !atomic.CompareAndSwapInt32(&tp.Running, 0, 1) {
		return nil
	}

	tp.Wg.Add(1)
	procValue, err := json.Marshal(tp.TaskProc)
	if err != nil {
		return err
	}

	_, err = etcdclient.GetEtcdClient().PutWithTTL(tp.Key(), string(procValue), config.GetConfig().System.TaskProcTtl)
	if err != nil {
		return err
	}

	tp.Wg.Done()
	return nil
}

func WatchTaskProc(NodeUUID string) clientv3.WatchChan {
	return etcdclient.GetEtcdClient().Watch(fmt.Sprintf(etcdclient.KeyEtcdNodeProcPrefix, NodeUUID), clientv3.WithPrefix())
}





