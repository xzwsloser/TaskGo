package service

import (
	"fmt"
	"sync"

	"github.com/xzwsloser/TaskGo/model"
	"github.com/xzwsloser/TaskGo/pkg/etcdclient"
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

func (nw *NodeWatcherService) GetTasksCount(nodeUUID string) (int, error) {
	resp, err := nw.client.Get(fmt.Sprintf(etcdclient.KeyEtcdTaskPrefix, nodeUUID),
				 clientv3.WithPrefix(), clientv3.WithCountOnly())
	if err != nil {
		return 0, err
	}
	return int(resp.Count), nil
}









