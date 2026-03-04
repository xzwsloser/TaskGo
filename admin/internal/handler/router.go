package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xzwsloser/TaskGo/admin/internal/middleware"
)

func configRouter(r *gin.Engine) {
	health := r.Group("/ping")
	{
		health.GET("", func(c *gin.Context) {
			c.JSON(http.StatusOK, "200")
		})
	}

	base := r.Group("")
	{
		base.POST("login", userHandler.Login)
		base.POST("register", userHandler.Register)
	}

	node := r.Group("node")
	node.Use(middleware.JWTAuth())
	{
		node.POST("delete", nodeHandler.Delete)
		node.POST("search", nodeHandler.Search)
	}

	script := r.Group("/script")
	script.Use(middleware.JWTAuth())
	{
		script.POST("add", scriptHandler.CreateOrUpdate)
		script.POST("delete", scriptHandler.Delete)
		script.POST("find", scriptHandler.FindById)
		script.POST("search", scriptHandler.Search)
	}
}
