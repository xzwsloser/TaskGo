package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/xzwsloser/TaskGo/model"
	"github.com/xzwsloser/TaskGo/node/server/handler"
	"github.com/xzwsloser/TaskGo/pkg/etcdclient"
	"github.com/xzwsloser/TaskGo/pkg/logger"
	"github.com/xzwsloser/TaskGo/pkg/utils"
	"go.etcd.io/etcd/api/v3/mvccpb"
)

var (
	ErrKeyMarshalFailCreate	= errors.New("Failed To Marshal Key While Create")
	ErrKeyMarshalFailModify = errors.New("Failed To Marshal Key While Modify")
)

func (srv *NodeServer) watchTasks() {
	rch := handler.WatchTasks(srv.UUID)
	for resp := range rch {
		for _, ev := range resp.Events {
			switch {
			case ev.IsCreate():
				// Task Create
				var task handler.Task
				if err := json.Unmarshal(ev.Kv.Value, &task); err != nil {
					err = ErrKeyMarshalFailCreate
					continue
				}
				srv.tasks[task.ID] = &task
				task.AssociatedNodeInfo(model.TaskStatusAssigned, srv.UUID, srv.Hostname, srv.IP)
				srv.addTask(&task)
			case ev.IsModify():
				var task handler.Task
				if err := json.Unmarshal(ev.Kv.Value, &task); err != nil {
					err = ErrKeyMarshalFailModify
					continue
				}
				task.AssociatedNodeInfo(model.TaskStatusAssigned, srv.UUID, srv.Hostname, srv.IP)
				srv.modifyTask(&task)
			case ev.Type == mvccpb.DELETE:
				srv.deleteTask(handler.GetTaskFromKey(string(ev.Kv.Key)))
			default:
				logger.GetLogger().Warn(fmt.Sprintf("Watch UnKnown Event Type [%v] From Task [%s]", ev.Type, string(ev.Kv.Key)))
			}
		}
	}
}

func getUUIDFromKey(key string) string {
	idx := strings.LastIndex(key, "/")
	if idx == -1 {
		return ""
	}
	return key[idx+1:]
}

func (srv *NodeServer) watchSystemInfo() {
	rch := handler.WatchSystem(srv.UUID)
	for  resp := range rch {
		for _, ev := range resp.Events {
			switch {
			case ev.IsCreate() || ev.IsModify():
				key := string(ev.Kv.Key)
				if string(ev.Kv.Value) != model.NodeSystemInfoSwitch || srv.UUID != getUUIDFromKey(key) {
					logger.GetLogger().Error(fmt.Sprintf("Get System Info From Node [%s] Not Alive", srv.UUID))
					continue
				}

				serverInfo, err := utils.GetSystemInfo()
				if err != nil {
					logger.GetLogger().Error(fmt.Sprintf("Failed To Get System From Info Node [%s] Failed: %v",srv.UUID, err))
					continue
				}
				
				buf, err := json.Marshal(serverInfo)
				if err != nil {
					logger.GetLogger().Error(fmt.Sprintf("Failed To Marshal System Info Of Node [%s] Failed: %v", srv.UUID, err))
					continue
				}

				_, err = etcdclient.GetEtcdClient().
									PutWithTTL(
									fmt.Sprintf(etcdclient.KeyEtcdSystemGet, 
									getUUIDFromKey(key)), 
									string(buf), 5*60)
				if err != nil {
					logger.GetLogger().Error(fmt.Sprintf("Put System Info From Node [%s] Failed: %v", srv.UUID, err))
				}
			}
		}
	}
}

// @Description: Find The Task To Exec Immediately
func (srv *NodeServer) watchOnce() {
	rch := handler.WatchOnce()
	for resp := range rch {
		for _, ev := range resp.Events {
			switch {
			case ev.IsModify(), ev.IsCreate():
				if len(ev.Kv.Value) != 0 && string(ev.Kv.Value) != srv.UUID {
					continue
				}

				task, ok := srv.tasks[handler.GetTaskFromKey(string(ev.Kv.Key))]
				if !ok {
					continue
				}

				go task.RunWithRecover()
			}
		}
	}
}


