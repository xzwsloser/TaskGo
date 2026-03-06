package service

import (
	"errors"
	"fmt"

	"github.com/xzwsloser/TaskGo/model"
	"gorm.io/gorm"
)

type UserService struct {
}

func (s *UserService) Login(username, password string) (*model.User, error) {
	u := &model.User{}
	u.UserName = username
	u.Password = password
	err := u.FindPartInfo()
	return u, err
}

func (s *UserService) FindByUsername(username string) (*model.User, error) {
	u := &model.User{}
	u.UserName = username
	err := u.FindByUsername()
	if err != nil {
		fmt.Printf("[Debug] Error = %s\n", err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound){
			return nil, nil
		} 
		return nil, err
	}

	return u, err
}



