package resp

import "github.com/xzwsloser/TaskGo/model"

type (
	RspNodeSearch struct {
		model.Node
		TaskCount int `json:"task_count"`
	}
)


