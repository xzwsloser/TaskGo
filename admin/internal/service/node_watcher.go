package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/xzwsloser/TaskGo/model"
	"github.com/xzwsloser/TaskGo/pkg/config"
	"github.com/xzwsloser/TaskGo/pkg/etcdclient"
	"github.com/xzwsloser/TaskGo/pkg/logger"
	"github.com/xzwsloser/TaskGo/pkg/notify"
	"github.com/xzwsloser/TaskGo/pkg/utils"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap/buffer"
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

// @Description: Get Task List Of Certain Node
func (nw *NodeWatcherService) GetTasksOfNode(nodeUUID string) ([]model.Task, error) {
	resp, err := nw.client.Get(fmt.Sprintf(etcdclient.KeyEtcdTaskPrefix, nodeUUID), clientv3.WithPrefix())
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("failed to get task list of node [%s]", nodeUUID))
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		logger.GetLogger().Warn(fmt.Sprintf("node %s have no task", nodeUUID))
		return nil, nil
	}

	tasks := make([]model.Task, 0)
	for _, pair := range resp.Kvs {
		var task model.Task
		if err = json.Unmarshal([]byte(pair.Value), &task); err != nil {
			logger.GetLogger().Error(fmt.Sprintf("failed to Unmarshal task %s from node %s, error: %s", 
				string(pair.Key), nodeUUID, err.Error()))
			continue
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// @Descrption: When Node Failed, Call This Function To Transport Task
func (nw *NodeWatcherService) FailOver(nodeUUID string) ([]int, []int, error) {
	tasks, err := nw.GetTasksOfNode(nodeUUID)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("node [%s] get task failed: %s", nodeUUID, err.Error()))
		return nil, nil, err
	}

	success := make([]int, 0)
	fail	:= make([]int, 0)
	taskService := &TaskService{}

	for _, task := range tasks {
		if task.Type == model.TaskTypeCmd && !config.GetConfig().System.CmdAutoAllocation {
			logger.GetLogger().Warn(
				fmt.Sprintf("task [%d] is cmd task on node [%s] is cmd task cannot transport",
			    task.ID, nodeUUID))
			fail = append(fail, task.ID)
			continue
		}

		oldUUID := task.RunOn
		allocatedUUID := taskService.AutoAllocateNode()
		if allocatedUUID == "" {
			logger.GetLogger().Warn(fmt.Sprintf("task [%d] on node [%s] cannot find a node to run", task.ID, oldUUID))
			fail = append(fail, task.ID)
			continue
		}

		err := nw.assignTask(allocatedUUID, &task)
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("task [%d] on node [%s] assign failed: %s", task.ID, allocatedUUID, err.Error()))
			fail = append(fail, task.ID)
			continue
		}

		_, err = nw.client.Delete(fmt.Sprintf(etcdclient.KeyEtcdTaskFormat, oldUUID, task.ID))
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("task [%d] on node [%s] failed to delete: %s", task.ID, oldUUID, err.Error()))
			fail = append(fail, task.ID)
			continue
		}
		success = append(success, task.ID)
	}
	return success, fail, nil
}

func transArrayToStr(arr []int) string {
	buf := &buffer.Buffer{}
	buf.WriteString("[")
	for idx, value := range arr {
		buf.WriteString(strconv.Itoa(value))
		if idx != len(arr)-1 {
			buf.WriteString(",")
		}
	}
	buf.WriteString("]")
	return buf.String()
}

// @Description: Watcher For Node Status
func (nw *NodeWatcherService) watcher() {
	rch := nw.client.Watch(etcdclient.KeyEtcdNodePrefix, clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case mvccpb.PUT:
				// Add Node
				nw.setNodeList(nw.getUUID(string(ev.Kv.Key)), string(ev.Kv.Value))
			case mvccpb.DELETE:
				// Delete Node
				uuid := nw.getUUID(string(ev.Kv.Key))
				nw.delFromNodeList(uuid)
				logger.GetLogger().Warn(fmt.Sprintf("task node [%s] delete event", uuid))
				node := &model.Node{UUID: uuid}
				err := node.FindByUUID()
				if err != nil {
					logger.GetLogger().Error(fmt.Sprintf("task node [%s] can not find, err: %s", uuid, err.Error()))
					continue
				}

				success, fail, err := nw.FailOver(uuid)
				if err != nil {
					logger.GetLogger().Error(fmt.Sprintf("task node [%s] fail over failed: %s",uuid, err.Error()))
					continue
				}

				if len(fail) == 0 {
					err = node.Delete()
					if err != nil {
						logger.GetLogger().Error(fmt.Sprintf("task node [%s] failed to delete: %s", uuid, err.Error()))
					}
				}
				msg := &notify.Message{
					Type: notify.NotifyEmail,
					IP: fmt.Sprintf("%s:%s", node.IP, node.PID),
					Subject: "TaskGo 节点失活报警",
					Body: fmt.Sprintf("[TaskGo Warning] TaskGo Node [%s] In Cluster Failed, Fail Over Success Count: %d TaskID are: %s, Fail Count: %d TaskID are: %s", 
											uuid, 
											len(success), 
											transArrayToStr(success), 
											len(fail), 
											transArrayToStr(fail)),
					To: config.GetConfig().Email.To,
					OccurTime: time.Now().Format(utils.TimeFormatSecond),
				}

				go notify.Send(msg)
			}
		}
	}
}

// @Description: Load exists Node And Start watcher Goroutine
func (nw *NodeWatcherService) Watch() error {
	_, err := nw.extractNodes()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("extrace node error: %s", err.Error()))
		return err
	}

	go nw.watcher()
	return nil
}





