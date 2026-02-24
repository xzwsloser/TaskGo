package model

import (
	"encoding/json"
	"sync"
	"time"
)

type TaskProcVal struct {
	Time	time.Time	`json:"time"`   	// Time Begin to Exec
	Killed	bool		`json:"killed"`		// Is Task Killed 
}

type TaskProc struct {
	ID			int		`json:"id"`
	TaskID		int		`json:"task_id"`
	NodeUUID	string	`json:"node_uuid"`
	TaskProcVal

	Running		int32
	Wg			sync.WaitGroup
}

func (tp *TaskProc) Val() (string, error) {
	tpv, err := json.Marshal(&tp.TaskProcVal)
	if err != nil {
		return "", err
	}

	return string(tpv), nil
}




