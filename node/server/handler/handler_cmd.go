package handler

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/xzwsloser/TaskGo/model"
	"github.com/xzwsloser/TaskGo/pkg/logger"
)

type CMDHandler struct {
}

func (c *CMDHandler) Run(task *Task) (result string, err error) {
	var (
		cmd  *exec.Cmd
		proc *TaskProc
	)

	if task.Timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(task.Timeout)*time.Second)
		defer cancel()
		cmd = exec.CommandContext(ctx, task.Cmd[0], task.Cmd[1:]...)
	} else {
		cmd = exec.Command(task.Cmd[0], task.Cmd[1:]...)
	}

	buffer := &bytes.Buffer{}
	cmd.Stdout = buffer
	cmd.Stderr = buffer

	err = cmd.Start()
	result = buffer.String()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("%s\n%s", result, err.Error()))
		return 
	}

	proc = &TaskProc{
		TaskProc: &model.TaskProc{
			ID: cmd.Process.Pid,
			TaskID: task.ID,
			NodeUUID: task.RunOn,
			TaskProcVal: model.TaskProcVal{
				Time: time.Now(),
				Killed: false,
			},
		},
	}

	// Register In Etcd
	err = proc.Start()
	defer proc.Stop()

	if err = cmd.Wait() ; err != nil {
		logger.GetLogger().Error(fmt.Sprintf("%s\n%s", result, err.Error()))
		return result, err
	}

	return result, nil
}

func RunPresetScript(script *model.Script) (result string, err error) {
	var cmd *exec.Cmd
	cmd = exec.Command(script.Cmd[0], script.Cmd[1:]...)
	buffer := &bytes.Buffer{}
	cmd.Stdout = buffer
	cmd.Stderr = buffer

	err = cmd.Start()
	result = buffer.String()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("Run Preset Script Error. %s\n%s", result, err.Error()))
		return 
	}

	if err = cmd.Wait() ; err != nil {
		logger.GetLogger().Error(fmt.Sprintf("Run Preset Script Error. %s\n%s", result, err.Error()))
		return 
	}

	return result, nil
}

