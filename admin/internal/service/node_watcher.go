package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/xzwsloser/TaskGo/model"
	"github.com/xzwsloser/TaskGo/pkg/config"
	"github.com/xzwsloser/TaskGo/pkg/etcdclient"
	"github.com/xzwsloser/TaskGo/pkg/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type NodeWatcherService struct {
	client		*etcdclient.EtcdClient
	nodeLists	map[string]model.Node
	mu			sync.Mutex
}

var nodeWatcherService *NodeWatcherService 

func NewNodeWatcherService() *NodeWatcherService {
	nodeWatcherService = &NodeWatcherService{
		client: etcdclient.GetEtcdClient(),
		nodeLists: make(map[string]model.Node),
	}

	return nodeWatcherService
}

func GetNodeWatcherService() *NodeWatcherService {
	return nodeWatcherService
}

// @Description: Count The All Task
func (nw *NodeWatcherService) GetTasksCount(nodeUUID string) (int, error) {
	resp, err := nw.client.Get(fmt.Sprintf(etcdclient.KeyEtcdTaskPrefix, nodeUUID),
				 clientv3.WithPrefix(), clientv3.WithCountOnly())
	if err != nil {
		return 0, err
	}
	return int(resp.Count), nil
}

// @Description: Transform Node List To Node UUID List
func (nw *NodeWatcherService) NodeListToArray() []string {
	nw.mu.Lock()
	defer nw.mu.Unlock()
	nodes := make([]string, 0)

	for k := range nw.nodeLists {
		nodes = append(nodes, k)
	}

	return nodes 
}

// @Description: Assign Task To Certain Node
func (nw *NodeWatcherService) assignTask(nodeUUID string, task *model.Task) error {
	if nodeUUID == "" {
		return fmt.Errorf("node uuid cannot be null")
	}
	node, ok := nw.nodeLists[nodeUUID]
	if !ok {
		return fmt.Errorf("assign unassigned task [%d] but node [%s] not exists", task.ID, node.UUID)
	}
	task.AssociatedNodeInfo(model.TaskStatusAssigned, node.UUID, node.Hostname, node.IP)

	taskBuf, err := json.Marshal(task)
	if err != nil {
		return err
	}
	_, err = etcdclient.GetEtcdClient().Put(fmt.Sprintf(etcdclient.KeyEtcdTaskFormat, nodeUUID, task.ID), string(taskBuf), )
	if err != nil {
		return err
	}

	err = task.Update()
	if err != nil {
		return err
	}

	return nil
}

// @Description: Add Node To List And Assign UnAssigned Task
func (nw *NodeWatcherService) setNodeList(key, val string) {
	var node model.Node
	err := json.Unmarshal([]byte(val), &node)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("discover node [%s] json error: %s", key, err.Error()))
	}
	nw.mu.Lock()
	nw.nodeLists[key] = node
	nw.mu.Unlock()

	// wait for new node to start
	time.Sleep(5 * time.Second)

	taskService := &TaskService{}
	tasks, err := taskService.GetNotAssignTasks()
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("discover node [%s], pid [%s] and get not assigned task err: %v", 
				key, val, err.Error()))
		return 
	}

	// assign task
	for _, task := range tasks {
		if task.Type == model.TaskTypeCmd && !config.GetConfig().System.CmdAutoAllocation {
			logger.GetLogger().Warn(fmt.Sprintf("assign task [%d] can not support assign", task.ID))
			continue
		}

		err = task.Unmarshal()
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("assign unassigned task [%d] marshal failed: %s", task.ID, err.Error()))
			continue
		}
		oldUUID := task.RunOn
		nodeUUID  := taskService.AutoAllocateNode()
		if nodeUUID == "" {
			nodeUUID = key
		}

		err = nw.assignTask(nodeUUID, &task)
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("assign unassigned task [%d] error: %s", task.ID, err.Error()))
			continue
		}

		_, err = etcdclient.GetEtcdClient().Delete(fmt.Sprintf(etcdclient.KeyEtcdTaskFormat, oldUUID, task.ID))
		if err != nil {
			logger.GetLogger().
				   Error(fmt.Sprintf("node [%s] task [%d] fail over delete key failed: %s", 
						 nodeUUID, task.ID, err.Error()))
			continue
		}
	}
} 

func (nw *NodeWatcherService) getUUID(key string) string {
	idx := strings.LastIndex(key , "/")
	if idx < 0 {
		return ""
	}

	return key[idx+1:]
}

// @Description: Load Node Registerd In Etcd To Node List
func (nw *NodeWatcherService) extractNodes() ([]string, error) {
	resp, err := etcdclient.GetEtcdClient().Get(etcdclient.KeyEtcdNodePrefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	nodes := make([]string, 0)
	if resp == nil || resp.Kvs == nil {
		return nodes, nil
	}

	for idx := range resp.Kvs {
		if val := resp.Kvs[idx].Value; val != nil {
			// Add Node Into List And Assign Task
			nw.setNodeList(nw.getUUID(string(resp.Kvs[idx].Key)), string(resp.Kvs[idx].Value))
			nodes = append(nodes, string(val))
		}
	}
	
	return nodes, nil
}

// @Description: Delete Failed Node From Node List
func (nw *NodeWatcherService) delFromNodeList(key string) {
	nw.mu.Lock()
	defer nw.mu.Unlock()
	delete(nw.nodeLists, key)
	logger.GetLogger().Info(fmt.Sprintf("delete node [%s] from nodelist", key))
}














