package request

import (
	"encoding/json"

	"github.com/xzwsloser/TaskGo/model"
)

type (
	ReqTaskUpdate struct {
			*model.Task
			Allocation int `json:"allocation" form:"allocation" binding:"required"`
	}

	ReqTaskSearch struct {
		PageInfo
		ID     int            	`json:"id" form:"id"`
		Name   string         	`json:"name" form:"name"`
		RunOn  string         	`json:"run_on" form:"run_on"`
		Type   model.TaskType 	`json:"task_type" form:"type"`
		Status int            	`json:"status" form:"status"`
	}

	ReqTaskLogSearch struct {
		PageInfo
		Name     string `json:"name" form:"name"`
		TaskId	 int    `json:"task_id" form:"task_id"`
		NodeUUID string `json:"node_uuid" form:"node_uuid"`
		Success  bool   `json:"success" form:"success"`
	}

	ReqTaskOnce struct {
		TaskId	 int    `json:"task_id" form:"task_id"`
		NodeUUID string `json:"node_uuid" form:"node_uuid"`
	}
)

func (r *ReqTaskUpdate) Valid() error {
	// default automatic assignment
	if r.Allocation == 0 {
		r.Allocation = model.AutoAllocation
	}
	notifyTo, _ := json.Marshal(r.NotifyToArray)
	r.NotifyTo = notifyTo
	scriptID, _ := json.Marshal(r.ScriptIDArray)
	r.ScriptID = scriptID
	return r.Check()
}
