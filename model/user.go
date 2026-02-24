package model

import "github.com/xzwsloser/TaskGo/pkg/dbclient"

const (
	RoleNormal	= 1
	RoleAdmin	= 2

	TaskGoUserTableName = "user"
)

type User struct {
	ID       int    `json:"id" gorm:"column:id;primary_key;auto_increment"`
	UserName string `json:"username" gorm:"size:128;column:username;not null"`
	Password string `json:"password" gorm:"size:128;column:password;not null"`
	Email    string `json:"email" gorm:"size:64;column:email;default:''"`
	Role     int    `json:"role" gorm:"size:1;column:role;default:1"`

	Created int64 `json:"created" gorm:"column:created;not null"`
	Updated int64 `json:"updated" gorm:"column:updated;default:0"`
}

func (u *User) TableName() string {
	return TaskGoUserTableName
}

func (u *User) Update() error {
	return dbclient.GetMysqlDB().Table(u.TableName()).Updates(u).Error
}

func (u *User) Delete() error {
	return dbclient.GetMysqlDB().Table(u.TableName()).Delete(u).Error
}

func (u *User) Insert() (int, error) {
	err := dbclient.GetMysqlDB().Table(u.TableName()).Create(u).Error
	if err != nil {
		return -1, err
	}
	return u.ID, nil
}

func (u *User) FindById() error {
	return dbclient.GetMysqlDB().Table(u.TableName()).
		Select("id", "username", "email", "role", "created", "updated").
		First(u).Error
}



