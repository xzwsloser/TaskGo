package service

import (
	"errors"
	"fmt"
	"github.com/xzwsloser/TaskGo/admin/internal/model/request"
	"github.com/xzwsloser/TaskGo/model"
	"github.com/xzwsloser/TaskGo/pkg/etcdclient"
	"github.com/xzwsloser/TaskGo/pkg/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type NodeService struct {
}

var (
	ErrNodeRunning		= errors.New("Can not Delete a Running Node")
)

func (*NodeService) DeleteNode(uuid string) error {
	node := &model.Node{UUID: uuid}
	err := node.FindByUUID()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("failed to find node: %v", err))
		return err
	}

	if node.Status == model.NodeConnSuccess {
		return ErrNodeRunning
	}

	// Delete Associated Task
	_, _ = etcdclient.GetEtcdClient().
			Delete(fmt.Sprintf(etcdclient.KeyEtcdTaskPrefix, node.UUID),
			clientv3.WithPrefix())
	err = node.Delete()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("failed to delete node: %v", err))
		return err
	}

	return nil
}

func (*NodeService) Search(r *request.ReqNodeSearch) ([]model.Node, int64, error) {
	page, pageSize := r.Page, r.PageSize
	node := &model.Node{}
	node.UUID = r.UUID
	node.IP = r.IP
	node.Status = r.Status
	node.UpTime = r.UpTime

	nodes, total, err := node.FindAndPage(page, pageSize)
	return nodes, total, err
}

func (*NodeService) GetNodeCount(status int) (int64, error) {
	node := &model.Node{}
	node.Status = status
	total, err := node.GetNodeCount()
	if err != nil {
		return 0, nil
	}
	return total, nil
}




