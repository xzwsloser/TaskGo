package config

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"github.com/xzwsloser/TaskGo/pkg/utils"
)

type (
	Mysql struct {
		Address 	 string `mapstructure:"address" json:"address" yaml:"address" ini:"address"`
		Port         string `mapstructure:"port" json:"port" yaml:"port" ini:"port"`
		Config       string `mapstructure:"config" json:"config" yaml:"config" ini:"config"`
		Dbname       string `mapstructure:"db-name" json:"dbname" yaml:"db-name" ini:"db-name"`
		Username     string `mapstructure:"username" json:"username" yaml:"username" ini:"username"`
		Password     string `mapstructure:"password" json:"password" yaml:"password" ini:"password"`
		MaxIdleConns int    `mapstructure:"max-idle-conns" json:"maxIdleConns" yaml:"max-idle-conns" ini:"max-idle-conns"`
		MaxOpenConns int    `mapstructure:"max-open-conns" json:"maxOpenConns" yaml:"max-open-conns" ini:"max-open-conns"`
		LogMode      string `mapstructure:"log-mode" json:"logMode" yaml:"log-mode" ini:"log-mode"`
		LogZap       bool   `mapstructure:"log-zap" json:"logZap" yaml:"log-zap" ini:"log-zap"`
	}

	Email struct {
		Port     int      `mapstructure:"port" json:"port" yaml:"port" ini:"port"`
		From     string   `mapstructure:"from" json:"from" yaml:"from" ini:"from"`
		Host     string   `mapstructure:"host" json:"host" yaml:"host" ini:"host"`
		IsSSL    bool     `mapstructure:"is-ssl" json:"isSSL" yaml:"is-ssl" ini:"is-ssl"`
		Secret   string   `mapstructure:"secret" json:"secret" yaml:"secret" ini:"secret"`
		Nickname string   `mapstructure:"nickname" json:"nickname" yaml:"nickname" ini:"nickname"`
		To       []string `mapstructure:"to" json:"to" yaml:"to" ini:"to"`
	}

	Etcd struct {
		Endpoints   []string `mapstructure:"endpoints" json:"endpoints" yaml:"endpoints" ini:"endpoints"`
		Username    string   `mapstructure:"username" json:"username" yaml:"username" ini:"username"`
		Password    string   `mapstructure:"password" json:"password" yaml:"password" ini:"password"`
		DialTimeout int64    `mapstructure:"dial-timeout" json:"dial-timeout" yaml:"dial-timeout" ini:"dial-timeout"`
		ReqTimeout  int64    `mapstructure:"req-timeout" json:"req-timeout" yaml:"req-timeout" ini:"req-timeout"`
	}

	System struct {
		Env                string `mapstructure:"env" json:"env" yaml:"env" ini:"env"`
		Addr               int    `mapstructure:"addr" json:"addr" yaml:"addr" ini:"addr"`
		NodeTtl            int64  `mapstructure:"node-ttl" json:"node-ttl" yaml:"node-ttl" ini:"node-ttl"`
		JobProcTtl         int64  `mapstructure:"job-proc-ttl" json:"job-proc-ttl" yaml:"job-proc-ttl" ini:"job-proc-ttl"`
		Version            string `mapstructure:"version" json:"version" yaml:"version" ini:"version"`
		LogCleanPeriod     int64  `mapstructure:"log-clean-period" json:"log-clean-period" yaml:"log-clean-period" ini:"log-clean-period"`
		LogCleanExpiration int64  `mapstructure:"log-clean-expiration" json:"log-clean-expiration" yaml:"log-clean-expiration" ini:"log-clean-expiration"`
		CmdAutoAllocation  bool   `mapstructure:"cmd-auto-allocation" json:"cmd-auto-allocation" yaml:"cmd-auto-allocation" ini:"cmd-auto-allocation"`
	}

	Log struct {
		Level         string `mapstructure:"level" json:"level" yaml:"level" ini:"level"`
		Format        string `mapstructure:"format" json:"format" yaml:"format" ini:"format"`
		Prefix        string `mapstructure:"prefix" json:"prefix" yaml:"prefix" ini:"prefix"`
		Director      string `mapstructure:"director" json:"director"  yaml:"director" ini:"director"`
		ShowLine      bool   `mapstructure:"show-line" json:"showLine" yaml:"showLine" ini:"showLine"`
		EncodeLevel   string `mapstructure:"encode-level" json:"encodeLevel" yaml:"encode-level" ini:"encode-level"`
		StacktraceKey string `mapstructure:"stacktrace-key" json:"stacktraceKey" yaml:"stacktrace-key" ini:"stacktrace-key"`
		LogInConsole  bool   `mapstructure:"log-in-console" json:"logInConsole" yaml:"log-in-console" ini:"log-in-console"`
	}

	Config struct {
		Log     Log     `mapstructure:"log" json:"log" yaml:"log" ini:"log"`
		Email   Email   `mapstructure:"email" json:"email" yaml:"email" ini:"email"`
		System  System  `mapstructure:"system" json:"system" yaml:"system" ini:"system"`
		Mysql   Mysql   `mapstructure:"mysql" json:"mysql" yaml:"mysql" ini:"mysql"`
		Etcd    Etcd    `mapstructure:"etcd" json:"etcd" yaml:"etcd" ini:"etcd"`
	}
)

var _defaultConfig *Config

func LoadConfig(configFileName string) (*Config, error) {
	var c Config
	v := viper.New()
	v.SetConfigFile(configFileName)
	ext := utils.GetExtOfFile(configFileName)
	v.SetConfigType(ext)
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Failed to read config file: %s", err))
	}
	v.WatchConfig()
	v.OnConfigChange(func (e fsnotify.Event) {
		fmt.Printf("config file changed: %s", e.Name)
		if err := v.Unmarshal(&c) ; err != nil {
			fmt.Println("error while change config")
		}
	})

	err = v.Unmarshal(&c)
	if err != nil {
		panic(fmt.Errorf("Failed to parse config file: %s", err))
	}

	_defaultConfig = &c
	fmt.Printf("the config you use is: \n%v\n", c)
	return &c, nil
}

func GetConfig() *Config {
	if _defaultConfig == nil {
		panic("Filed to get config")
	}

	return _defaultConfig
}

