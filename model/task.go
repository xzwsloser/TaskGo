package model

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/xzwsloser/TaskGo/pkg/dbclient"
	"github.com/xzwsloser/TaskGo/pkg/logger"
	"github.com/xzwsloser/TaskGo/pkg/utils"
)

type taskType int

const (
	// task type 
	taskTypeCmd  = taskType(1)
	taskTypeHttp = taskType(2)

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
	Type          taskType `json:"task_type" gorm:"size:1;column:type;not null;" binding:"required"`
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

	if len(t.Cmd) == 0 && t.Type == taskTypeCmd {
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





