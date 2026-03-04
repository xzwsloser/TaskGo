package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
	{
		node.POST("delete", nodeHandler.Delete)
		node.POST("search", nodeHandler.Search)
	}
}
