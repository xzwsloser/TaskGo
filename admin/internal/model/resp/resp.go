package resp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
	@Description: Rule The Format Of Resp
*/
type (
	PageResult struct {
		List     any		 `json:"list"`
		Total    int64       `json:"total"`
		Page     int         `json:"page"`
		PageSize int         `json:"page_size"`
	}

	Response struct {
		Code 	int		`json:"code"`
		Data 	any		`json:"data"`
		Msg		string	`json:"msg"`
	}
)

const (
	SUCCESS		= 200
	ERROR		= 1000

	ErrorRequestParameter	= 1001
	ErrorTaskFormat			= 1002
	ErrorTokenGenerate		= 1003
	ErrorUserNameExist		= 1004
)

func Result(code int, data any, msg string, c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Data: data,
		Msg:  msg,
	})
}

func Ok(c *gin.Context) {
	Result(SUCCESS, map[string]any{}, "success", c)
}

func OkWithMessage(msg string, c *gin.Context) {
	Result(SUCCESS, map[string]any{}, msg, c)
}

func OkWithData(data any, msg string, c *gin.Context) {
	Result(SUCCESS, data, msg, c)
}

func OkWithDetailed(data any, msg string, c *gin.Context) {
	Result(SUCCESS, data, msg, c)
}

func FailWithMessage(code int, msg string, c *gin.Context) {
	Result(code, map[string]any{}, msg, c)
}

func FailWithCode(code int, c *gin.Context) {
	Result(code, map[string]any{}, "failed", c)
}

func FailWithDetailed(code int, data any, msg string, c *gin.Context) {
	Result(code, data, msg, c)
}

