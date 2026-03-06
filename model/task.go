package model

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/xzwsloser/TaskGo/pkg/dbclient"
	"github.com/xzwsloser/TaskGo/pkg/logger"
	"github.com/xzwsloser/TaskGo/pkg/utils"
)

type TaskType int

const (
	// task type 
	TaskTypeCmd  = TaskType(1)
	TaskTypeHttp = TaskType(2)

	// http method
	HttpMethodGet  = 1
	HttpMethodPost = 2

	// task exec status
	TaskExecSuccess = 1
	TaskExecFail	= 2

	// task assign status
	TaskStatusNotAssigned = 1
	TaskStatusAssigned 	  = 2

	// allocation ways
	ManualAllocation = 1
	AutoAllocation   = 2

	// tablename
	TaskGoTaskTableName = "task"
)

var (
	ErrEmptyTaskName = errors.New("Task Name is empty.")
	ErrEmptyTaskCmd	 = errors.New("Task Cmd is empty.")
)

// register to  /taskGo/task/<node_uuid>/<task_id>
type Task struct {
	ID      int    `json:"id" gorm:"column:id;primary_key;auto_increment"`
	Name    string `json:"name" gorm:"size:64;column:name;not null;index:idx_task_name" binding:"required"`
	Command string `json:"command" gorm:"type:text;column:command;not null" binding:"required"`
	//preset script ID
	ScriptID      []byte `json:"-"  gorm:"size:256;column:script_id;default:null"`
	ScriptIDArray []int  `json:"script_id" gorm:"-"`
	//Timeout setting of task execution time, which is effective when it is greater than 0.
	Timeout int64 `json:"timeout" gorm:"size:13;column:timeout;default:0"`
	// Retry times of task execution failures
	// The default value is 0
	RetryTimes int `json:"retry_times" gorm:"size:4,column:retry_times;default:0"`
	// Retry interval for task execution failure
	// in seconds. If the value is less than 0, try again immediately
	RetryInterval int64    `json:"retry_interval" gorm:"size:10;column:retry_interval;default:0"`
	Type          TaskType `json:"task_type" gorm:"size:1;column:type;not null;" binding:"required"`
	HttpMethod    int     `json:"http_method" gorm:"size:1;column:http_method"`
	NotifyType    int     `json:"notify_type" gorm:"size:1;column:notify_type;not null"`
	// Whether to allocate nodes
	Status        int    `json:"status" gorm:"size:1;column:status;not null;default:0;index:idx_task_status"`
	NotifyTo      []byte `json:"-" gorm:"size:256;column:notify_to;default:null"`
	NotifyToArray []int  `json:"notify_to" gorm:"-"`
	Spec          string `json:"spec" gorm:"size:64;column:spec;not null"`
	RunOn         string `json:"run_on" gorm:"size:128;column:run_on;index:idx_task_run_on;"`
	Note          string `json:"note" gorm:"size:512;column:note;default:''"`
	Created       int64  `json:"created" gorm:"column:created;not null"`
	Updated       int64  `json:"updated" gorm:"column:updated;default:0"`

	Hostname string   `json:"host_name" gorm:"-"`
	Ip       string   `json:"ip" gorm:"-"`
	Cmd      []string `json:"cmd" gorm:"-"`
}

type DateCount struct {
	Date	string	`json:"date" gorm:"column:date"`
	Count	string	`json:"count" gorm:"column:count"`
}

func (t *Task) TableName() string {
	return TaskGoTaskTableName
}

func (t *Task) AssociatedNodeInfo(status int, nodeUUID, hostName, ip string) {
	t.Status, t.RunOn, t.Hostname, t.Ip = status, nodeUUID, hostName, ip
}

func (t *Task) Insert() (int, error) {
	err := dbclient.GetMysqlDB().Table(t.TableName()).Create(t).Error
	if err != nil {
		return -1, err
	}

	return t.ID, nil
}

func (t *Task) Update() error {
	return dbclient.GetMysqlDB().Table(t.TableName()).Updates(t).Error
}

func (t *Task) Delete() error {
	return dbclient.GetMysqlDB().Table(t.TableName()).Delete(t).Error
}

func (t *Task) FindById() error {
	return dbclient.GetMysqlDB().Table(t.TableName()).
		Where("id = ?", t.ID).First(t).Error
}

func (t *Task) UpdateByNodeUUID(nodeUUID string) error {
	return dbclient.GetMysqlDB().Table(t.TableName()).
		Where("run_on = ?", t.RunOn).Updates(t).Error
}

func (t *Task) SplitCmd() {
	commands := strings.SplitN(t.Command, " ", 2)
	if len(commands) == 1 {
		t.Cmd = commands
		return
	}

	t.Cmd = make([]string, 0, 2)
	t.Cmd = append(t.Cmd, commands[0])
	t.Cmd = append(t.Cmd, utils.ParseCmdArguments(commands[1])...)
}

func (t *Task) Check() error {
	t.Name = strings.TrimSpace(t.Name)
	if len(t.Name) == 0 {
		return ErrEmptyTaskName
	}
	
	if t.RetryInterval == 0 {
		t.RetryTimes = 1
	}

	if len(strings.TrimSpace(t.Command)) == 0 {
		return ErrEmptyTaskCmd
	}

	if len(t.Cmd) == 0 && t.Type == TaskTypeCmd {
		t.SplitCmd()
	}

	return nil
}

func (t *Task) Val() string {
	data, err := json.Marshal(t)
	if err != nil {
		logger.GetLogger().Error(err.Error())
		return ""
	}

	return string(data)
}

func (t *Task) Unmarshal() error {
	err := json.Unmarshal(t.NotifyTo, &t.NotifyToArray)
	if err != nil {
		logger.GetLogger().Error(err.Error())
		return err
	}

	err =  json.Unmarshal(t.ScriptID, &t.ScriptIDArray)
	if err != nil {
		logger.GetLogger().Error(err.Error())
	}
	return err
}

func (t *Task) FindAndPage(page int, pageSize int) ([]Task, int64, error) {
	db := dbclient.GetMysqlDB().Table(t.TableName())
	if t.ID > 0 {
		db = db.Where("id = ?", t.ID)
	}

	if len(t.Name) > 0 {
		db = db.Where("name like ?", t.Name + "%")
	}

	if len(t.RunOn) > 0 {
		db = db.Where("run_on = ?", t.RunOn)
	}

	if t.Type > 0 {
		db = db.Where("type = ?", t.Type)
	}

	if t.Status > 0 {
		db = db.Where("status = ?", t.Status)
	}

	var total int64
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	tasks := make([]Task, 0, 2)
	err = db.Limit(pageSize).Offset((page-1)*pageSize).Find(&tasks).Error
	if err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

func (t *Task) GetNotAssignedTasks() ([]Task, error) {
	var tasks []Task
	err := dbclient.GetMysqlDB().
				Table(t.TableName()).
				Where("status = ?", TaskStatusNotAssigned).
				Find(&tasks).Error
	return tasks, err
}

/* func (t *Task) GetTaskCountAfter(startTime int64, success int) (int64, error) {
	db := dbclient.GetMysqlDB().
				   Table(t.TableName()).
				   Where("start_time > ? and end_time != 0 and success = ?", startTime, success)
	var total int64
	err := db.Count(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
} */

/* func (t *Task) GetTaskExecCount(start, end int64, success int) ([]DateCount, error) {
	var dateCount []DateCount
	db := dbclient.GetMysqlDB().Table(t.TableName()).
			Select("FROM_UNIXTIME( start_time, '%Y-%m-%d' ) AS date", "COUNT( * ) AS count").
			Group("date").
			Order("date ASC").
			Where("start_time > ? and start_time < ? and end_time != 0 and success = ?", start, end, success)
	err := db.Find(&dateCount).Error
	if err != nil {
		return nil, err
	}
	return dateCount, nil
} */

