package etcdclient

import (
	"context"

	"github.com/xzwsloser/TaskGo/pkg/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// service register using etcd
type Register struct {
	client 			*EtcdClient
	stop   			chan struct{}	
	keepAliveChan	<-chan *clientv3.LeaseKeepAliveResponse
	cancel			context.CancelFunc
	leaseID			clientv3.LeaseID
	ttl				int64
}

func NewRegister(ttl int64) *Register {
	return &Register{
		client: GetEtcdClient(),
		stop: make(chan struct{}),
		ttl: ttl,
	}
} 

func (r *Register) RegisterService(key, val string) error {
	err := r.grant()
	if err != nil {
		logger.GetLogger().Error(err.Error())
		return err
	}

	_, err = r.client.Put(key, val, clientv3.WithLease(r.leaseID))
	if err != nil {
		logger.GetLogger().Error(err.Error())
		return err
	}

	go r.keepAlive(key, val)

	return nil
}

func (r *Register) grant() error {

	// Application Lease ID
	leaseResp, err := r.client.Grant(r.ttl)
	if err != nil {
		logger.GetLogger().Error(err.Error())
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	keepAliveChan, err := r.client.KeepAlive(ctx, leaseResp.ID)

	if err != nil {
		logger.GetLogger().Error(err.Error())
		return err
	}

	r.leaseID = leaseResp.ID
	r.cancel = cancel
	r.keepAliveChan = keepAliveChan

	return nil
}

func (r *Register) keepAlive(key, value string) {
	for {
		select {
		case <-r.stop:
			// unregister
			err := r.RevokeLease()
			if err != nil {
				logger.GetLogger().Error(err.Error())
				return 
			}
			return

		case _, ok := <- r.keepAliveChan:
			if !ok {
				logger.GetLogger().Info("The Lease ID Expired.")

				// register the key again
				err := r.RegisterService(key, value)
				if err != nil {
					logger.GetLogger().Error(err.Error())
				}

				return 
			}
		}
	}
}

func (r *Register) Stop() {
	r.stop <- struct{}{}
}

func (r *Register) RevokeLease() error {
	r.cancel()
	_, err := r.client.Revoke(r.leaseID)
	return err
}

