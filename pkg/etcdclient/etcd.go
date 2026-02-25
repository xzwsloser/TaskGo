package etcdclient

import ( "context"
	"errors"
	"fmt"
	"time"

	"github.com/xzwsloser/TaskGo/pkg/config"
	"github.com/xzwsloser/TaskGo/pkg/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdClient struct {
	client		*clientv3.Client
	reqTimeout	time.Duration
}

var _defaultEtcdClient *EtcdClient

var (
	ErrKeyMayChanged = errors.New("Key in Etcd may changed.")
)

func InitEtcdClient() (*EtcdClient, error) {
	ec := config.GetConfig().Etcd
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: ec.Endpoints,
		DialTimeout: time.Duration(ec.ReqTimeout) * time.Second,
	})
	if err != nil {
		fmt.Printf("Failed to init Etcd Client.\n")
		return nil, err
	}

	_defaultEtcdClient = &EtcdClient{
		client: cli,
		reqTimeout: time.Duration(ec.ReqTimeout),
	}

	return _defaultEtcdClient, nil
}

func GetEtcdClient() *EtcdClient {
	if _defaultEtcdClient == nil {
		logger.GetLogger().Error("Etcd Client is not Init.")
		return nil
	}
	return _defaultEtcdClient
}

// grant detailed err info
type etcdTimeoutCtx struct {
	context.Context
	endpoints []string
}

func (c *etcdTimeoutCtx) Err() error {
	err := c.Context.Err()
	if err == context.DeadlineExceeded {
		err = fmt.Errorf("%s: etcd(%v) error\n", err, c.endpoints)
	}
	return err
}

func NewEtcdTimeoutCtx(c *EtcdClient) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), c.reqTimeout)
	etcdCtx := &etcdTimeoutCtx{}
	etcdCtx.Context = ctx
	etcdCtx.endpoints = config.GetConfig().Etcd.Endpoints
	return etcdCtx, cancel
}

func (cli *EtcdClient) Put(key, val string, 
	options ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	ctx, cancel := NewEtcdTimeoutCtx(cli)
	defer cancel()
	return cli.client.Put(ctx, key, val, options...)
}

func (cli *EtcdClient) PutWithTTL(key, val string, ttl int64, 
	options ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	ctx, cancel := NewEtcdTimeoutCtx(cli)
	defer cancel()
	leaseResp, err := cli.Grant(ttl)
	if err != nil {
		logger.GetLogger().Error(err.Error())
		return nil, err
	}

	options = append(options, clientv3.WithLease(leaseResp.ID))
	return cli.client.Put(ctx, key, val, options...)
}

// put k-v into etcd using positivie lock
func (cli *EtcdClient) PutWithModRev(key, val string, rev int64) (*clientv3.PutResponse, error) {
	ctx, cancel := NewEtcdTimeoutCtx(cli)

	// transcation
	resp, err := cli.client.Txn(ctx).
		If(clientv3.Compare(clientv3.ModRevision(key), "=", rev)).
		Then(clientv3.OpPut(key, val)).Commit()
	cancel()

	if err != nil {
		logger.GetLogger().Error(err.Error())
		return nil, err
	}

	if !resp.Succeeded {
		logger.GetLogger().Error("key in etcd may changed.")
		return nil, ErrKeyMayChanged
	}

	putResp := clientv3.PutResponse(*resp.Responses[0].GetResponsePut())
	return &putResp, nil
}


func (cli *EtcdClient) Grant(ttl int64) (*clientv3.LeaseGrantResponse, error) {
	ctx, cancel := NewEtcdTimeoutCtx(cli)
	defer cancel()
	return cli.client.Grant(ctx, ttl)
}

func (cli *EtcdClient) Get(key string, 
	options ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	ctx, cancel := NewEtcdTimeoutCtx(cli)
	defer cancel()
	return cli.client.Get(ctx, key, options...)
}

func (cli *EtcdClient) Delete(key string,
	options ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	ctx, cancel := NewEtcdTimeoutCtx(cli)
	defer cancel()
	return cli.client.Delete(ctx, key, options...)
}

func (cli *EtcdClient) Watch(key string,
	options ...clientv3.OpOption) clientv3.WatchChan {
	return cli.client.Watch(context.Background(), key, options...)
}

// delete lease ID and keys binded to it 
func (cli *EtcdClient) Revoke(leaseID clientv3.LeaseID) (*clientv3.LeaseRevokeResponse, error) {
	ctx, cancel := NewEtcdTimeoutCtx(cli)
	defer cancel()
	return cli.client.Revoke(ctx, leaseID)
}

func (cli *EtcdClient) KeepAliveOnce(leaseID clientv3.LeaseID) (*clientv3.LeaseKeepAliveResponse, error) {
	ctx, cancel := NewEtcdTimeoutCtx(cli)
	defer cancel()
	return cli.client.KeepAliveOnce(ctx, leaseID)
}

// Dstribute Lock
func (cli *EtcdClient) GetLock(key string, leaseID clientv3.LeaseID) (bool, error) {
	lockKey := fmt.Sprintf(KeyEtcdLockFormat, key)
	ctx, cancel := NewEtcdTimeoutCtx(cli)
	resp, err := cli.client.Txn(ctx).
		If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).
		Then(clientv3.OpPut(key, "", clientv3.WithLease(leaseID))).
		Commit()
	if err != nil {
		logger.GetLogger().Error(err.Error())
		return false, err
	}
	cancel()

	return resp.Succeeded, nil
}

func (cli *EtcdClient) DeleteLock(key string) error {
	lockKey := fmt.Sprintf(KeyEtcdLockFormat, key)
	_, err := cli.Delete(lockKey)
	return err
}

