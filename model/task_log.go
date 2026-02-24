package model

import "github.com/xzwsloser/TaskGo/pkg/dbclient"

const (
	TaskGoTaskLogTableName = "task_log"
)

type TaskLog struct {
	ID       int    `json:"id" gorm:"column:id;primary_key;auto_increment"`
	Name     string `json:"name" gorm:"size:64;column:name;index:idx_task_log_name;not null"`
	TaskId	 int    `json:"task_id" gorm:"column:task_id;index:idx_task_log_id; not null"`
	Command  string `json:"command" gorm:"size:512;column:command"`
	IP       string `json:"ip" gorm:"size:32;column:ip"` // node ip
	Hostname string `json:"hostname" gorm:"size:32;column:hostname"`
	NodeUUID string `json:"node_uuid" gorm:"size:128;column:node_uuid;not null;index:idx_task_log_node"`
	Success  bool   `json:"success" gorm:"size:1;column:success;not null"`

	Output string `json:"output" gorm:"size:512;column:output;"`
	Spec   string `json:"spec" gorm:"size:64;column:spec;not null" `

	RetryTimes int   `json:"retry_times" gorm:"size:4;column:retry_times;default:0"`
	StartTime  int64 `json:"start_time" gorm:"column:start_time;not null;"`
	EndTime    int64 `json:"end_time" gorm:"column:end_time;default:0;"`
}

func (tl *TaskLog) TableName() string {
	return TaskGoTaskLogTableName
}

func (tl *TaskLog) Update() error {
	return dbclient.GetMysqlDB().Table(tl.TableName()).Updates(tl).Error
}

func (tl *TaskLog) Delete() error {
	return dbclient.GetMysqlDB().Table(tl.TableName()).Delete(tl).Error
}

func (tl *TaskLog) Insert() (int, error) {
	err := dbclient.GetMysqlDB().Table(tl.TableName()).Create(tl).Error
	if err != nil {
		return -1, err
	}

	return tl.ID, nil
}




