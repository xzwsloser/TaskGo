package notify

import (
	"strings"
	"time"

	"github.com/xzwsloser/TaskGo/pkg/config"
	"github.com/xzwsloser/TaskGo/pkg/utils"
)

type Message struct {
	Type		int
	IP			string
	Subject		string
	Body		string
	To			[]string
	OccurTime	string
}

var msgQue chan *Message

func InitNoticer() {
	mailConfig := config.GetConfig().Email
	_defaultMail = &Mail{
		Port: mailConfig.Port,
		From: mailConfig.From,
		Host: mailConfig.Host,
		Secret: mailConfig.Secret,
		Nickname: mailConfig.Nickname,
	}

	msgQue = make(chan *Message, 64)
}

func Send(msg *Message) {
	msgQue <- msg
}

func Serve() {
	for {
		select {
		case msg, ok := <- msgQue:
			if !ok {
				return 
			}

			if msg == nil {
				continue
			}

			msg.Check()
			_defaultMail.SendMsg(msg)
		}
	}
}

func (m *Message) Check() {
	if m.OccurTime == "" {
		m.OccurTime = time.Now().Format(utils.TimeFormatSecond)
	}
	//Remove the transfer character
	m.Body = strings.ReplaceAll(m.Body, "\"", "'")
	m.Body = strings.ReplaceAll(m.Body, "\n", "")
}
