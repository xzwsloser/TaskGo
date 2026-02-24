package model

import (
	"errors"
	"strings"

	"github.com/xzwsloser/TaskGo/pkg/dbclient"
	"github.com/xzwsloser/TaskGo/pkg/utils"
)

const (
	TaskGoScriptTableName	= "script"
)

var (
	ErrEmptyScriptName	= errors.New("Script Name is empty.")
	ErrEmptyScriptCmd	= errors.New("Script Cmd is empty.")
)

//Preset Script
type Script struct {
	ID      int    `json:"id" gorm:"column:id;primary_key;auto_increment"`
	Name    string `json:"name" gorm:"size:256;column:name;not null;index:idx_script_name" binding:"required"`
	Command string `json:"command" gorm:"type:text;column:command;not null" binding:"required"`
	Created int64  `json:"created" gorm:"column:created;not null"`
	Updated int64  `json:"updated" gorm:"column:updated;default:0"`

	Cmd []string `json:"cmd" gorm:"-"`
}

func (s *Script) TableName() string {
	return TaskGoScriptTableName
}

func (s *Script) Insert() (int, error) {
	err := dbclient.GetMysqlDB().Table(s.TableName()).Create(s).Error
	if err != nil {
		return -1, err
	}
	return s.ID, nil
}

func (s *Script) Update() error {
	return dbclient.GetMysqlDB().Table(s.TableName()).Updates(s).Error
}

func (s *Script) Delete() error {
	return dbclient.GetMysqlDB().Table(s.TableName()).Delete(s).Error
}

func (s *Script) FindById() error {
	return dbclient.GetMysqlDB().Table(s.TableName()).
		Where("id = ?", s.ID).First(s).Error
}

func (s *Script) SplitCmd() {
	commands := strings.SplitN(s.Command, " ", 2)
	if len(commands) == 1 {
		s.Cmd = commands
		return 
	}

	s.Cmd = make([]string, 0, 2)
	s.Cmd = append(s.Cmd, commands[0])
	s.Cmd = append(s.Cmd, utils.ParseCmdArguments(commands[1])...)
}

func (s *Script) Check() error {
	s.Name = strings.TrimSpace(s.Name)
	if len(s.Name) == 0 {
		return ErrEmptyScriptName
	}

	if len(strings.TrimSpace(s.Command)) == 0 {
		return ErrEmptyScriptCmd
	}

	if len(s.Cmd) == 0 {
		s.SplitCmd()
	}

	return nil
}







