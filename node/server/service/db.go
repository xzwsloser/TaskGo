package service

import (
	"github.com/xzwsloser/TaskGo/model"
	"github.com/xzwsloser/TaskGo/pkg/dbclient"
)

func RegisterTables() {
	_ = dbclient.GetMysqlDB().AutoMigrate(
		model.User{},
		model.Node{},
		model.Task{},
		model.TaskLog{},
    )
}
