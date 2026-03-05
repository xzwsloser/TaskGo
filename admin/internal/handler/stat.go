package handler

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xzwsloser/TaskGo/admin/internal/model/request"
	"github.com/xzwsloser/TaskGo/admin/internal/model/resp"
	"github.com/xzwsloser/TaskGo/admin/internal/service"
	"github.com/xzwsloser/TaskGo/model"
	"github.com/xzwsloser/TaskGo/pkg/etcdclient"
	"github.com/xzwsloser/TaskGo/pkg/logger"
	"github.com/xzwsloser/TaskGo/pkg/utils"
)

type StatHandler struct {
}

var (
	statHandler	= new(StatHandler)
)

const (
	WeekUnix	= 60*60*24*7
)

// @Router: /stat/today
// @Method: GET
// @Description: Get Today Statistics Of Node Server
func (*StatHandler) GetTodayStatistics(c *gin.Context) {
	taskExcSuccess, err := taskService.GetTodayTaskExecCount(model.TaskExecSuccess)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("[get_statistics] get success task failed: %s", err.Error()))
	}

	taskExcFail, err := taskService.GetTodayTaskExecCount(model.TaskExecFail)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("[get_statistics] get fail task failed: %s", err.Error()))
	}

	taskRunningCount, err := taskService.GetRunningTaskCount()
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("[get_statistics] get running task count failed: %s", err.Error()))
	}

	normalNodeCount, err := nodeService.GetNodeCount(model.NodeConnSuccess)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("[get_statistics] get success node count error: %s", err.Error()))
	}

	failNodeCount, err := nodeService.GetNodeCount(model.NodeConnFail)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("[get_statistics] get fail node count error: %s", err.Error()))
	}

	resp.OkWithDetailed(resp.RspSystemStatisics{
		NormalNodeCount: normalNodeCount,
		FailNodeCount: failNodeCount,
		TaskExcSuccessCount: taskExcSuccess,
		TaskExcFailCount: taskExcFail,
		TaskRunningCount: taskRunningCount,
	}, "ok", c)
}

// @Router: /stat/week
// @Method: GET
// @Description: Get Week Statistics Of Node
func (*StatHandler) GetWeekStatistics(c *gin.Context) {
	t := time.Now()
	tasksExcSuccess, err := taskService.GetTaskExecCount(t.Unix()-WeekUnix, t.Unix(), model.TaskExecSuccess)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("[get_week_statistic] get success tasks failed: %s", err.Error()))
	}

	tasksExcFail, err := taskService.GetTaskExecCount(t.Unix()-WeekUnix, t.Unix(), model.TaskExecFail)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("[get_week_statistic] get fail tasks failed: %s", err.Error()))
	}

	resp.OkWithDetailed(resp.RspDateCountSet{
		SuccessDateCount: tasksExcSuccess,
		FailDateCount: tasksExcFail,
	}, "ok", c)
}

// @Router: /stat/system
// @Method: GET
// @Description: Get The System Info Of Certain Server
func (*StatHandler) GetSystemInfo(c *gin.Context) {
	var req request.ByUUID
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[get_system_info] request parameter error:%s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[get_system_info] request parameter error", c)
		return
	}
	var server *utils.SystemInfo
	var err error
	if req.UUID == "" {
		//Get native information of admin
		server, err = utils.GetSystemInfo()
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("[get_system_info]  error:%s", err.Error()))
			resp.FailWithMessage(resp.ERROR, "[get_system_info]  error", c)
			return
		}
	} else {
		//Set the survival time to 30 seconds
		_, err := etcdclient.GetEtcdClient().PutWithTTL(fmt.Sprintf(etcdclient.KeyEtcdSystemSwitch, req.UUID), model.NodeSystemInfoSwitch, 30)
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("get system info from node[%s] etcd put error: %s", req.UUID, err.Error()))
			resp.FailWithMessage(resp.ERROR, "[get_system_info]  error", c)
			return
		}
		//There will be a delay. By default, wait 2s.
		time.Sleep(2 * time.Second)
		server, err = service.GetNodeWatcherService().GetNodeSystemInfo(req.UUID)
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("get system info from node[%s] watch key error: %s", req.UUID, err.Error()))
			resp.FailWithMessage(resp.ERROR, "[get_system_info]  error", c)
			return
		}
	}
	resp.OkWithDetailed(server, "ok", c)
}


