package config

import "testing"

func TestLoadConfig(t *testing.T) {
	configPath := "../../node/conf/config.json"
	c, err := LoadConfig(configPath)
	if err != nil {
		t.Error("Failed to fetch config")
	}
	t.Logf("mysql: %v\n", c.Mysql)
	t.Logf("etcd:  %v\n", c.Etcd)
	t.Logf("email: %v\n", c.Email)
	t.Logf("system: %v\n", c.System)
	t.Logf("log: %v\n", c.Log)
}

