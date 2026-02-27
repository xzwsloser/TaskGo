package service

import (
	"fmt"
	"net/http"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/xzwsloser/TaskGo/pkg/config"
	"github.com/xzwsloser/TaskGo/pkg/dbclient"
	"github.com/xzwsloser/TaskGo/pkg/etcdclient"
	"github.com/xzwsloser/TaskGo/pkg/logger"
	"github.com/xzwsloser/TaskGo/pkg/notify"
	"github.com/xzwsloser/TaskGo/pkg/utils"
)

/*
	Init Node Server (EtcdClient、MySQL、Logger ...)
*/

const (
	Version = "v1.0.0"
)

var (
	NodeOptions struct {
		flags.Options
		Version        bool   `short:"v" long:"verbose"  description:"Show nodeServer version"`
		EnablePProfile bool   `short:"p" long:"enable-pprof"  description:"enable pprof"`
		PProfilePort   int    `short:"d" long:"pprof-port"  description:"pprof port" default:"8188"`
		ConfigFileName string `short:"c" long:"config" description:"Use nodeServer config file" default:"main"`
		EnableDevMode  bool   `short:"m" long:"enable-dev-mode"  description:"enable dev mode"`
	}
)

func InitNodeServer() (*config.Config, error) {
	var parser = flags.NewParser(&NodeOptions, flags.Default)
	if _, err := parser.Parse() ; err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		return nil, err
	}

	if NodeOptions.Version {
		fmt.Printf("TaskGo Node Server Version: %s\n", Version)
		os.Exit(0)
	}

	if NodeOptions.EnablePProfile {
		fmt.Printf("enable pprof http server at: %d\n", NodeOptions.PProfilePort)
		go func () {
			http.ListenAndServe(fmt.Sprintf(":%d", NodeOptions.PProfilePort), nil)
		}()
	}

	if !utils.IsFileExists(NodeOptions.ConfigFileName) {
		fmt.Printf("Config File Not Exisits")
		os.Exit(0)
	}

	defaultConfig, err := config.LoadConfig(NodeOptions.ConfigFileName)
	if err != nil {
		fmt.Printf("Failed to get config: %s\n", err.Error())
		os.Exit(0)
	}

	// log
	l := logger.InitLogger("TaskGo")
	if l == nil {
		fmt.Printf("Failed to Init Logger")
		os.Exit(0)
	}

	// mysql
	createSql := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 ;", 
		config.GetConfig().Mysql.Dbname)
	err = dbclient.CreateDatabase("mysql", createSql)
	if err != nil {
		fmt.Printf("Failed to Create Database.\n")
		os.Exit(0)
	}

	_, err = dbclient.InitMysql()
	if err != nil {
		fmt.Printf("Failed to Init Mysql")
		os.Exit(0)
	} else {
		fmt.Printf("Connect to Mysql at %s:%s\n", config.GetConfig().Mysql.Address, 
				config.GetConfig().Mysql.Port)
	}

	// email
	notify.InitNoticer()

	// etcdclient
	_, err = etcdclient.InitEtcdClient()
	if err != nil {
		fmt.Printf("Failed to Init Etcd Client")
		os.Exit(0)
	} else {
		fmt.Printf("Connect to Etcd at %v\n", config.GetConfig().Etcd.Endpoints)
	}

	return defaultConfig, nil
}


