package main

import (
	"fmt"
	"os"

	"github.com/xzwsloser/TaskGo/admin/internal/handler"
	"github.com/xzwsloser/TaskGo/admin/internal/server"
	"github.com/xzwsloser/TaskGo/admin/internal/service"
	"github.com/xzwsloser/TaskGo/pkg/logger"
	"github.com/xzwsloser/TaskGo/pkg/notify"
)

func main() {
	srv, err := server.NewApiServer()
	if err != nil {
		fmt.Printf("failed to create api server, error: %s\n", err.Error())
		os.Exit(1)
	}
	// register handler
	srv.RegisterRouters(handler.RegisterHandlers)

	// start node watcher
	_ = service.NewNodeWatcherService()
	err = service.GetNodeWatcherService().Watch()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("watch node failed: %s", err.Error()))
		os.Exit(1)
	}

	// init tables
	err = service.RegisterTables()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("failed to register table %s", err.Error()))
		os.Exit(1)
	}

	// notify operation
	go notify.Serve()

	// start api server
	err = srv.ListenAndServe()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("start api server error: %s", err.Error()))
		os.Exit(1)
	}

	os.Exit(0)
}

