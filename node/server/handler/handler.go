package handler

import "github.com/xzwsloser/TaskGo/model"

type Handler interface {
	Run(*Task) (string, error)
}

func NewHandler(task *Task) Handler {
	var h Handler
	if task.Type == model.TaskTypeCmd {
		h = &CMDHandler{}
	} else {
		h = &HttpHandler{}
	}

	return h
}


