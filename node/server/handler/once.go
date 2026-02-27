package handler

import (
	"github.com/xzwsloser/TaskGo/pkg/etcdclient"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// @Description: Execute The Task Immediately
// Value
// If a single node is executed, value is nodeUUID
// If the node where the task is located needs to be executed, is null
func WatchOnce() clientv3.WatchChan {
	return etcdclient.GetEtcdClient().Watch(etcdclient.KeyEtcdOncePrefix, clientv3.WithPrefix())
}

