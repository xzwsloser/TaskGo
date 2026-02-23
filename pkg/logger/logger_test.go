package logger

import (
	"testing"

	"github.com/xzwsloser/TaskGo/pkg/config"
)

func TestLogger(t *testing.T) {
	configPath := "../../node/conf/config.json"
	config.LoadConfig(configPath)
	InitLogger("taskGo") 
	GetLogger().Debug("debug info")
	GetLogger().Info("info info")
	GetLogger().Warn("warn info")
	GetLogger().Error("error info")
}

