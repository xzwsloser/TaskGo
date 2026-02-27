package handler

import (
	"fmt"

	"github.com/xzwsloser/TaskGo/pkg/etcdclient"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func WatchSystem(nodeUUID string) clientv3.WatchChan {
	return etcdclient.GetEtcdClient().Watch(fmt.Sprintf(etcdclient.KeyEtcdSystemSwitch, nodeUUID), clientv3.WithPrefix())
}

