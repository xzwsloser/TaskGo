package handler

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xzwsloser/TaskGo/admin/internal/model/request"
	"github.com/xzwsloser/TaskGo/admin/internal/model/resp"
	"github.com/xzwsloser/TaskGo/admin/internal/service"
	"github.com/xzwsloser/TaskGo/model"
	"github.com/xzwsloser/TaskGo/pkg/config"
	"github.com/xzwsloser/TaskGo/pkg/etcdclient"
	"github.com/xzwsloser/TaskGo/pkg/logger"
)

type TaskHandler struct {
}

var (
	taskHandler	= new(TaskHandler)
	taskService = new(service.TaskService)
)

// @Router: /task/add
// @Method: POST
// @Description: Create Task And Allocate Node
func (*TaskHandler) CreateOrUpdate(c *gin.Context) {
	var req request.ReqTaskUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[create_task] request parameter error:%s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[create_task] request parameter error", c)
		return
	}
	if err := req.Valid(); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("create_task check error:%s", err.Error()))
		resp.FailWithMessage(resp.ErrorTaskFormat, "[create_task] check error", c)
		return
	}

	var err error
	var insertId int
	t := time.Now()

	if req.Allocation == model.AutoAllocation {
		if !config.GetConfig().System.CmdAutoAllocation && req.Type == model.TaskTypeCmd {
			resp.FailWithMessage(resp.ERROR, "[create_task] The shell command is not supported to automatically assign nodes by default.", c)
			return
		}
		// Automatic allocation
		nodeUUID := taskService.AutoAllocateNode()
		if nodeUUID == "" {
			logger.GetLogger().Error("[create_task] auto allocate node error")
			resp.FailWithMessage(resp.ERROR, "[create_task] auto allocate node error", c)
			return
		}
		req.RunOn = nodeUUID
	} else if req.Allocation == model.ManualAllocation {
		// Manual assignment
		if len(req.RunOn) == 0 {
			resp.FailWithMessage(resp.ERROR, "[create_task] manually assigned node can't be null", c)
			return
		}
		node := &model.Node{UUID: req.RunOn}
		_ = node.FindByUUID()
		if node.Status == model.NodeConnFail {
			resp.FailWithMessage(resp.ERROR, "[create_task] manually assigned node inactivation", c)
			return
		}
	}
	if req.ID > 0 {
		//update
		task := &model.Task{ID: req.ID}
		_ = task.FindById()
		oldNodeUUID := task.RunOn
		if oldNodeUUID != "" {
			_, err = etcdclient.GetEtcdClient().Delete(fmt.Sprintf(etcdclient.KeyEtcdTaskFormat, oldNodeUUID, req.ID))
			if err != nil {
				logger.GetLogger().Error(fmt.Sprintf("[update_task] delete etcd node[%s]  error:%s", oldNodeUUID, err.Error()))
				resp.FailWithMessage(resp.ERROR, "[update_task] delete etcd node error", c)
				return
			}
		}
		req.Updated = t.Unix()
		err = req.Update()
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("[update_task] into db  error:%s", err.Error()))
			resp.FailWithMessage(resp.ERROR, "[update_task] into db id error", c)
			return
		}
	} else {
		//create
		req.Created = t.Unix()
		insertId, err = req.Insert()
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("[create_task] insert task into db error:%s", err.Error()))
			resp.FailWithMessage(resp.ERROR, "[create_task] insert task into db error", c)
			return
		}
		req.ID = insertId
	}
	b, err := json.Marshal(req)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[create_task] json marshal task error:%s", err.Error()))
		resp.FailWithMessage(resp.ERROR, "[create_task] json marshal task error", c)
		return
	}
	// Allocate Task To Node
	_, err = etcdclient.GetEtcdClient().Put(fmt.Sprintf(etcdclient.KeyEtcdTaskFormat, req.RunOn, req.ID), string(b))
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[create_task] etcd put task error:%s", err.Error()))
		resp.FailWithMessage(resp.ERROR, "[create_task] etcd put task error", c)
		return
	}

	resp.OkWithDetailed(req, "operate success", c)
}

// @Router: /task/delete
// @Method: POST
// @Description: Delete Task By IDS
func (*TaskHandler) Delete(c *gin.Context) {
	var req request.ByIDS
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[delete_task] request parameter error:%s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[delete_task] request parameter error", c)
		return
	}
	for _, id := range req.IDS {
		task := model.Task{ID: id}
		err := task .FindById()
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("[delete_task] find task by id :%d error:%s", id, err.Error()))
			continue
		}
		_, err = etcdclient.GetEtcdClient().Delete(fmt.Sprintf(etcdclient.KeyEtcdTaskFormat, task.RunOn, id))
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("[delete_task] etcd delete task error:%s", err.Error()))
			continue
		}
		err = task.Delete()
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("[delete_task] into db error:%s", err.Error()))
			continue
		}
	}
	resp.OkWithMessage("delete success", c)
}

// @Router: /task/find
// @Method: GET
// @Description: Find Task By ID
func (*TaskHandler) FindById(c *gin.Context) {
	var req request.ByID
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[find_task] request parameter error:%s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[find_task] request parameter error", c)
		return
	}
	task := model.Task{ID: req.ID}
	err := task.FindById()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[find_task] find task by id :%d error:%s", req.ID, err.Error()))
		resp.FailWithMessage(resp.ERROR, "[find_task] find task by id error", c)
		return
	}
	if len(task.NotifyTo) != 0 {
		_ = task.Unmarshal()
	}
	resp.OkWithDetailed(task, "find success", c)
}

// @Router: /task/search
// @Method: POST
// @Description: Search For Task By Page And Condition
func (*TaskHandler) Search(c *gin.Context) {
	var req request.ReqTaskSearch
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[search_task] request parameter error:%s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[search_task] request parameter error", c)
		return
	}
	req.Check()
	tasks, total, err := taskService.Search(&req)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[search_task] search task error:%s", err.Error()))
		resp.FailWithMessage(resp.ERROR, "[search_task] search task error", c)
		return
	}
	var resultTasks []model.Task
	for _, task := range tasks {
		_ = task.Unmarshal()
		resultTasks = append(resultTasks, task)
	}
	resp.OkWithDetailed(resp.PageResult{
		List:     resultTasks,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, "search success", c)
}

// @Router: /task/search
// @Method: POST
// @Description: Search Task Log By Page And Condition
func (*TaskHandler) SearchTaskLog(c *gin.Context) {
	var req request.ReqTaskLogSearch
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[search_task_log] request parameter error:%s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[search_task_log] request parameter error", c)
		return
	}
	req.Check()
	tasks, total, err := taskService.SearchTaskLog(&req)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[search_task_log] db error:%s", err.Error()))
		resp.FailWithMessage(resp.ERROR, "[search_task_log] db error", c)
		return
	}
	resp.OkWithDetailed(resp.PageResult{
		List:     tasks,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, "search success", c)
}

// @Router: /task/once
// @Method: POST
// @Description: Exec Task Immediately
func (*TaskHandler) Once(c *gin.Context) {
	var req request.ReqTaskOnce
	var err error
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[task_once] request parameter error:%s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[task_once] request parameter error", c)
		return
	}
	//find node
	node := &model.Node{UUID: req.NodeUUID}
	err = node.FindByUUID()
	if err != nil || node.Status == model.NodeConnFail {
		logger.GetLogger().Error(fmt.Sprintf("[task_once] node[%s] conn fail:%v", req.NodeUUID, err))
		resp.FailWithMessage(resp.ERROR, "[task_once] node conn fail ", c)
		return
	}
	task := &model.Task{ID: req.TaskId}
	err = task.FindById()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[task_once] job_id[%d] not exist db:%s", req.TaskId, err.Error()))
		resp.FailWithMessage(resp.ERROR, "[task_once] job not exist ", c)
		return
	}

	err = taskService.ExecOnce(&req)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[task_once] etcd put job_id :%d error:%s", req.TaskId, err.Error()))
		resp.FailWithMessage(resp.ERROR, "[task_once] put  error", c)
		return
	}
	resp.OkWithMessage("task once success", c)
}





