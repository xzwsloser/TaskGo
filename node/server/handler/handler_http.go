package handler

import (
	"strings"
	"time"

	"github.com/xzwsloser/TaskGo/model"
	"github.com/xzwsloser/TaskGo/pkg/httpclient"
)

type HttpHandler struct {
}

const (
	HttpTaskTimeout int64	= 500
)

func (h *HttpHandler) Run(task *Task) (result string, err error) {
	var proc *TaskProc
	proc = &TaskProc{
		TaskProc: &model.TaskProc{
			ID: 0,
			TaskID: task.ID,
			NodeUUID: task.RunOn,
			TaskProcVal: model.TaskProcVal{
				Time: time.Now(),
			},
		},
	}

	err = proc.Start()
	if err != nil {
		return 
	}
	defer proc.Stop()

	if task.Timeout <= 0 || task.Timeout > HttpTaskTimeout {
		task.Timeout = HttpTaskTimeout
	}

	if task.HttpMethod == model.HttpMethodGet {
		result, err = httpclient.GetHttpClient().Get(task.Command, task.Timeout)
	} else {
		fields := strings.Split(task.Command, "?")
		url := fields[0]
		var body string
		if len(fields) >= 2 {
			body = fields[1]
		}

		result, err = httpclient.GetHttpClient().Post(url, body, task.Timeout)
	}

	return
}

