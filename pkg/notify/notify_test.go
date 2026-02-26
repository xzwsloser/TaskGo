package notify

import (
	"testing"
	"time"

	"github.com/xzwsloser/TaskGo/pkg/config"
	"github.com/xzwsloser/TaskGo/pkg/logger"
)

func TestNotify(t *testing.T) {
	configFile := "../../node/conf/config.json"
	config.LoadConfig(configFile)
	InitNoticer()
	logger.InitLogger("taskGo")
	msg := &Message{
		Type: 1,
		IP: "127.0.0.1",
		Subject: "Test Message",
		Body: "Hello World",
		To: []string{"3088655042@qq.com"},
		OccurTime: "",
	}

	go Serve()

	Send(msg)

	time.Sleep(time.Second * 10)
}

