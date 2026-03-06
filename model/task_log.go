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

func (tl *TaskLog) FindAndPage(page int, pageSize int) ([]TaskLog, int64, error) {
	db := dbclient.GetMysqlDB().Table(tl.TableName())
	if len(tl.Name) > 0 {
		db = db.Where("name like ?", tl.Name + "%")
	}

	if tl.TaskId > 0 {
		db = db.Where("task_id = ?", tl.TaskId)
	}

	if len(tl.NodeUUID) > 0 {
		db = db.Where("node_uuid = ?", tl.NodeUUID)
	}

	db = db.Where("success = ?", tl.Success)

	var total int64
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	taskLogs := make([]TaskLog, 0, 2)
	err = db.Limit(pageSize).Offset((page-1)*pageSize).Find(&taskLogs).Error
	if err != nil {
		return nil, 0, err
	}

	return taskLogs, total, err
}

func (tl *TaskLog) GetTaskCountAfter(startTime int64, success int) (int64, error) {
	db := dbclient.GetMysqlDB().
				   Table(tl.TableName()).
				   Where("start_time > ? and end_time != 0 and success = ?", startTime, success)
	var total int64
	err := db.Count(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}


func (tl *TaskLog) GetTaskExecCount(start, end int64, success int) ([]DateCount, error) {
	var dateCount []DateCount
	db := dbclient.GetMysqlDB().Table(tl.TableName()).
			Select("FROM_UNIXTIME( start_time, '%Y-%m-%d' ) AS date", "COUNT( * ) AS count").
			Group("date").
			Order("date ASC").
			Where("start_time > ? and start_time < ? and end_time != 0 and success = ?", start, end, success)
	err := db.Find(&dateCount).Error
	if err != nil {
		return nil, err
	}
	return dateCount, nil
}

