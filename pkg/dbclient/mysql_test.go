package dbclient

import (
	"testing"
	"github.com/xzwsloser/TaskGo/pkg/config"
)

func TestMysqlClient(t *testing.T) {
	configFile := "../../node/conf/config.json"
	config.LoadConfig(configFile)
	InitMysql()

	type testDBObj struct {
		Shorten		string	`json:"shorten" gorm:"column:shorten;primary_key"`
		Url			string	`json:"url" gorm:"column:url"`
	}

	var obj testDBObj
	err := GetMysqlDB().Table("shorturl").First(&obj).Error
	if err != nil {
		t.Error(err)
		return 
	}

	t.Logf("obj: \n%v\n", obj)
}

