package main

import (
	"fmt"
	"os"

	"github.com/xzwsloser/TaskGo/node/server/service"
	"github.com/xzwsloser/TaskGo/pkg/event"
	"github.com/xzwsloser/TaskGo/pkg/logger"
	"github.com/xzwsloser/TaskGo/pkg/notify"
)

func main() {
	// Load Config
	if _, err := service.InitNodeServer(); err != nil {
		fmt.Printf("Failed To Config Node Server: %s\n", err.Error())
		os.Exit(1)
	}

	// Init Node Server
	nodeServer, err := service.NewNodeServer()
	if err != nil {
		fmt.Printf("Failed To New Node Server: %s\n", err.Error())
		os.Exit(1)
	}

	service.RegisterTables()
	
	// Register Node Server
	if err = nodeServer.RegisterNodeServer(); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("Failed To Register Node Server: %s", err.Error()))
		os.Exit(1)
	}

	if err = nodeServer.Run(); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("Node Run Error: %s", err.Error()))
		os.Exit(1)
	}

	go notify.Serve()
	logger.GetLogger().Info(fmt.Sprintf("TaskGo Node %s Run ...", nodeServer.String()))

	// Grace Stop
	event.OnEvent(event.EXIT, nodeServer.Stop)
	event.WaitEvent()
	event.EmitEvent(event.EXIT, nil)
	logger.GetLogger().Info(fmt.Sprintf("TaskGo Node %s Exit Success", nodeServer.String()))
}


