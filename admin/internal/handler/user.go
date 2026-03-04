package handler

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xzwsloser/TaskGo/admin/internal/middleware"
	"github.com/xzwsloser/TaskGo/admin/internal/model/request"
	"github.com/xzwsloser/TaskGo/admin/internal/model/resp"
	"github.com/xzwsloser/TaskGo/admin/internal/service"
	"github.com/xzwsloser/TaskGo/model"
	"github.com/xzwsloser/TaskGo/pkg/logger"
	"github.com/xzwsloser/TaskGo/pkg/utils"
)

type UserHandler struct {
}

var (
	userHandler = new(UserHandler)
	userService = new(service.UserService)
)

// @Router: /login
// @Method: POST
// @Description: User Login
func (u *UserHandler) Login(c *gin.Context) {
	var req request.ReqUserLogin
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[user_login] request parameter error:%s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[user_login] request parameter error", c)
		return
	}
	user, err := userService.Login(req.UserName, req.Password)
	if err != nil || user.ID == 0 {
		logger.GetLogger().Error(fmt.Sprintf("[user_login] db error:%v", err))
		resp.FailWithMessage(resp.ERROR, "[user_login] username or password is incorrect", c)
		return
	}

	j := middleware.NewJWT()
	claims := j.CreateClaims(middleware.BaseClaims{
		ID:       user.ID,
		UserName: user.UserName,
	})
	token, err := j.CreateToken(claims)
	if err != nil {
		resp.FailWithMessage(resp.ErrorTokenGenerate, "获取token失败", c)
		return
	}
	resp.OkWithDetailed(resp.RspLogin{
		User:  user,
		Token: token,
	}, "login success", c)
}

// @Router: /register
// @Method: POST
// @Description: User Register
func (*UserHandler) Register(c *gin.Context) {
	var req request.ReqUserRegister
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[user_register] request parameter error:%s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[user_register] request parameter error", c)
		return
	}
	user, err := userService.FindByUsername(req.UserName)
	if user != nil || err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[user_register] db find by username:%s", req.UserName))
		resp.FailWithMessage(resp.ErrorUserNameExist, "[user_register] the user name has already been used", c)
		return
	}
	if req.Role == 0 {
		req.Role = model.RoleNormal
	}
	userModel := &model.User{
		UserName: req.UserName,
		Password: utils.MD5(req.Password),
		Role:     req.Role,
		Email:    req.Email,
		Created:  time.Now().Unix(),
	}
	insertId, err := userModel.Insert()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[user_register] db insert error:%v", err))
		resp.FailWithMessage(resp.ERROR, "[user_register] db insert error", c)
		return
	}
	userModel.ID = insertId
	resp.OkWithDetailed(userModel, "register success", c)
}

