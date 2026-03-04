package handler

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xzwsloser/TaskGo/admin/internal/model/request"
	"github.com/xzwsloser/TaskGo/admin/internal/model/resp"
	"github.com/xzwsloser/TaskGo/admin/internal/service"
	"github.com/xzwsloser/TaskGo/model"
	"github.com/xzwsloser/TaskGo/pkg/logger"
)

type ScriptHandler struct {
}

var (
	scriptHandler = new(ScriptHandler)
	scriptService = new(service.ScriptService)
)

// @Router: /script/add
// @Method: POST
// @Description: Create Or Update Preset Script
func (s *ScriptHandler) CreateOrUpdate(c *gin.Context) {
	var req model.Script
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[create_script] request parameter error:%s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[create_script] request parameter error", c)
		return
	}
	var err error
	t := time.Now()
	if req.ID > 0 {
		//update
		req.Updated = t.Unix()
		err = req.Update()
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("[update_script] into db  error:%s", err.Error()))
			resp.FailWithMessage(resp.ERROR, "[update_script] into db id error", c)
			return
		}
	} else {
		//create
		req.Created = t.Unix()
		_, err = req.Insert()
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("[create_script] insert script into db error:%s", err.Error()))
			resp.FailWithMessage(resp.ERROR, "[create_script] insert script into db error", c)
			return
		}
	}
	resp.OkWithDetailed(req, "operate success", c)
}

// @Router: /script/delete
// @Method: POST
// @Description: delete preset script
func (s *ScriptHandler) Delete(c *gin.Context) {
	var req request.ByIDS
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[delete_script] request parameter error:%s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[delete_script] request parameter error", c)
		return
	}
	for _, id := range req.IDS {
		script := model.Script{ID: id}
		err := script.FindById()
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("[delete_script] find script by id :%d error:%s", id, err.Error()))
			continue
		}
		err = script.Delete()
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("[delete_script] into db error:%s", err.Error()))
			continue
		}
	}
	resp.OkWithMessage("delete success", c)
}

// @Router: /script/find
// @Method: POST
// @Description: Find Preset Script By ID
func (s *ScriptHandler) FindById(c *gin.Context) {
	var req request.ByID
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[find_script] request parameter error:%s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[find_script] request parameter error", c)
		return
	}
	script := model.Script{ID: req.ID}
	err := script.FindById()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[find_script] find script by id :%d error:%s", req.ID, err.Error()))
		resp.FailWithMessage(resp.ERROR, "[find_script] find script by id error", c)
		return
	}
	resp.OkWithDetailed(script, "find success", c)
}

// @Router: /script/search
// @Method: POST
// @Description: Search Preset Script By Page And Condition
func (s *ScriptHandler) Search(c *gin.Context) {
	var req request.ReqScriptSearch
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[search_script] request parameter error:%s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[search_script] request parameter error", c)
		return
	}
	req.Check()
	scripts, total, err := scriptService.Search(&req)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[search_script] search script error:%s", err.Error()))
		resp.FailWithMessage(resp.ERROR, "[search_script] search script error", c)
		return
	}
	resp.OkWithDetailed(resp.PageResult{
		List:     scripts,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, "search success", c)
}
