package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jessevdk/go-flags"
	"github.com/xzwsloser/TaskGo/pkg/config"
	"github.com/xzwsloser/TaskGo/pkg/dbclient"
	"github.com/xzwsloser/TaskGo/pkg/etcdclient"
	"github.com/xzwsloser/TaskGo/pkg/logger"
	"github.com/xzwsloser/TaskGo/pkg/notify"
	"github.com/xzwsloser/TaskGo/pkg/utils"
)

const (
	shutdownMaxAge = 15 * time.Second
	shutdownWait   = 1000 * time.Millisecond

	AdminName	 	= "TaskGo/Admin"
	AdminVersion	= "v1.0.0"

	AdminProjectName		= "taskGo_Admin"
)


const (
	reset	= "\033[0m"
)

// Api Server Options
var (
	ApiOptions struct {
		flags.Options
		Version           bool   `short:"v" long:"verbose"  description:"Show ApiServer version"`
		EnablePProfile    bool   `short:"p" long:"enable-pprof"  description:"enable pprof"`
		PProfilePort      int    `short:"d" long:"pprof-port"  description:"pprof port" default:"8288"`
		EnableHealthCheck bool   `short:"a" long:"enable-health-check"  description:"enable health check"`
		HealthCheckURI    string `short:"i" long:"health-check-uri"  description:"health check uri" default:"/health" `
		HealthCheckPort   int    `short:"f" long:"health-check-port"  description:"health check port" default:"8186"`
		ConfigFileName    string `short:"c" long:"config" description:"Use ApiServer config file" default:"main"`
		EnableDevMode     bool   `short:"m" long:"enable-dev-mode"  description:"enable dev mode"`
	}
)

type ApiServer struct {
	Engine		*gin.Engine
	HttpServer	*http.Server
	Addr		string
	mu			sync.Mutex
	doneChan	chan struct{}

	// Register Routers
	Routers			[]func(*gin.Engine)
	// Middlewares 
	Middlewares		[]func(*gin.Engine)
	// Shutdown Callbacks (Hook) 
	Shutdowns		[]func(*ApiServer)
	Services		[]func(*ApiServer)
}


func (srv *ApiServer) getDoneChan() <-chan struct{} {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	return srv.getDoneChanLocked()
}

func (srv *ApiServer) getDoneChanLocked() chan struct{} {
	if srv.doneChan == nil {
		srv.doneChan = make(chan struct{})
	}
	return srv.doneChan
}

func (srv *ApiServer) Shutdown(ctx context.Context) {
	if len(srv.Shutdowns) > 0 {
		for _, shutdown := range srv.Shutdowns {
			shutdown(srv)
		}
	}

	<-time.After(shutdownMaxAge)

	srv.HttpServer.Shutdown(ctx)
}

// Api Server Recovery Middleware
func (srv *ApiServer) apiRecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				stack := utils.Stack(3)

				// Avoid Jwt Token Write Into Log
				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				headers := strings.Split(string(httpRequest), "\r\n")
				for idx, header := range headers {
					current := strings.Split(header, ":")
					if current[0] == "Authorization" {
						headers[idx] = current[0] + ": *"
					}
				}

				if brokenPipe {
					logger.GetLogger().Error(fmt.Sprintf("%s\n%s%s", err, string(httpRequest), reset))
				} else {
					logger.GetLogger().Error(fmt.Sprintf("[Recovery] %s panic recovered:\n%s\n%s%s",
						utils.FormatTime(time.Now()), err, stack, reset))
				}

				if brokenPipe {
					c.Error(err.(error))
					c.Abort()
				} else {
					c.AbortWithStatus(http.StatusInternalServerError)
				}
			}
		}()
		c.Next()
	}
}

func (srv *ApiServer) setupSignal() {
	go func() {
		var sigChan = make(chan os.Signal, 1)
		signal.Notify(sigChan, 
					  syscall.SIGINT, 
				      syscall.SIGHUP, 
				      syscall.SIGTERM)
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownMaxAge)
		defer shutdownCancel()

		// Receive OS Signal And Close Server
		for sig := range sigChan {
			if sig == syscall.SIGINT || 
			   sig == syscall.SIGHUP || 
			   sig == syscall.SIGTERM {
				logger.GetLogger().Error(fmt.Sprintf("Graceful shutdown:signal %v to stop api-server ", sig))
				srv.Shutdown(shutdownCtx)
			} else {
				logger.GetLogger().Info(fmt.Sprintf("Caught signal %v", sig))
			}
		}
		logger.Shutdown()
	}()
}

// @Description: Create Api Server
func NewApiServer(inits ...func()) (*ApiServer, error) {
	var parser = flags.NewParser(&ApiOptions, flags.Default)
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}

		return nil, err
	}

	if ApiOptions.Version {
		fmt.Printf("%s Version: %s\n", AdminName, AdminVersion)
		os.Exit(0)
	}

	if ApiOptions.EnablePProfile {
		go func() {
			fmt.Printf("enable ppro http server at: %d\n", ApiOptions.PProfilePort)
			fmt.Println(http.ListenAndServe(fmt.Sprintf(":%d", ApiOptions.PProfilePort), nil))
		}()
	}

	if ApiOptions.EnableHealthCheck {
		go func() {
			fmt.Printf("enable health check http server at: %d\n", ApiOptions.HealthCheckPort)
			fmt.Println(http.ListenAndServe(fmt.Sprintf(":%d", ApiOptions.HealthCheckPort), nil))
		}()
	}

	var configFile = ApiOptions.ConfigFileName
	defaultConfig, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Printf("api server failed to init config error: %s\n", err.Error())
		return nil, err
	}

	// logger
	logger.InitLogger(AdminProjectName)
	// notify
	notify.InitNoticer()

	// database
	mysqlConfig := config.GetConfig().Mysql
	createSql := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4;", mysqlConfig.Dbname)
	if err := dbclient.CreateDatabase("mysql", createSql); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("Create Mysql Database Failed, Error: %s", err.Error()))
	}
	_, err = dbclient.InitMysql()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("api-server:init mysql database failed: %s", err.Error()))
	} else {
		logger.GetLogger().Info("api-server:init mysql success")
	}

	// etcd
	_, err = etcdclient.InitEtcdClient()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("api-server:init etcd client failed: %s", err.Error()))
	}
	if len(inits) > 0 {
		for _, init := range inits {
			init()
		}
	}

	apiServer := &ApiServer{
		Addr: fmt.Sprintf(":%d", defaultConfig.System.Addr),
	}

	// Listen Os Signal For Grace Exit
	apiServer.setupSignal()

	return apiServer, nil
}


// @Description: Api Server Listen And Serve
func (srv *ApiServer) ListenAndServe() error {
	srv.Engine = gin.New()
	srv.Engine.Use(srv.apiRecoveryMiddleware())

	for _, service := range srv.Services {
		service(srv)
	}

	for _, middleware := range srv.Middlewares {
		middleware(srv.Engine)
	}

	for _, c := range srv.Routers {
		c(srv.Engine)
	}

	srv.HttpServer = &http.Server{
		Handler: srv.Engine,
		Addr: srv.Addr,
		ReadTimeout: 20*time.Second,
		WriteTimeout: 20*time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := srv.HttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}	

// Hook Function Register
func (srv *ApiServer) RegisterShutdown(handlers ...func(*ApiServer)) {
	srv.Shutdowns = append(srv.Shutdowns, handlers...)
}

func (srv *ApiServer) RegisterService(handlers ...func(*ApiServer)) {
	srv.Services = append(srv.Services, handlers...)
}

func (srv *ApiServer) RegisterMiddleware(middlewares...func(*gin.Engine)) {
	srv.Middlewares = append(srv.Middlewares, middlewares...)
}

func (srv *ApiServer) RegisterRouters(routers ...func(*gin.Engine)) {
	srv.Routers = append(srv.Routers, routers...)
}



