package service

import (
	"github.com/xzwsloser/TaskGo/model"
	"github.com/xzwsloser/TaskGo/pkg/dbclient"
)

func RegisterTables() error {
	db := dbclient.GetMysqlDB()
	err := db.AutoMigrate(
		model.User{},
		model.Task{},
		model.TaskLog{},
		model.Node{},
		model.Script{},
	)

	return err
}
