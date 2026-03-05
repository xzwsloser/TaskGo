package model

import (
	"github.com/xzwsloser/TaskGo/pkg/dbclient"
)

const (
	NodeConnSuccess    =  1
	NodeConnFail	   =  2

	TaskGoNodeTableName    = "node"

	NodeSystemInfoSwitch   = "alive"
)

// register to /taskGo/node/<node_uuid>/
type Node struct {
	ID       int    `json:"id" gorm:"column:id;primary_key;auto_increment"`
	PID      string `json:"pid" gorm:"size:16;column:pid;not null"`
	IP       string `json:"ip" gorm:"size:32;column:ip;default:''"`
	Hostname string `json:"hostname" gorm:"size:64;column:hostname;default:''"`
	UUID     string `json:"uuid" gorm:"size:128;column:uuid;not null;index:idx_node_uuid;"`
	Version  string `json:"version" gorm:"size:64;column:version;default:''"`
	Status   int    `json:"status" gorm:"size:1;column:status"`

	UpTime   int64 `json:"up" gorm:"column:up;not null"`
	DownTime int64 `json:"down" gorm:"column:down;default:0"`
}

func (n *Node) TableName() string {
	return TaskGoNodeTableName
}

func (n *Node) String() string {
	return "Node[" + n.UUID +"]" + " PID[" + n.PID + "]"
}

func (n *Node) Insert() (int, error) {
	err := dbclient.GetMysqlDB().Table(n.TableName()).Create(n).Error
	if err != nil {
		return -1, err
	}
	return n.ID, nil
}

func (n *Node) Update() error {
	return dbclient.GetMysqlDB().Table(n.TableName()).Updates(n).Error
}

func (n *Node) Delete() error {
	return dbclient.GetMysqlDB().Table(n.TableName()).
		Where("uuid = ?", n.UUID).Delete(&Node{}).Error
}

func (n *Node) FindByUUID() error {
	return dbclient.GetMysqlDB().Table(n.TableName()).
		Where("uuid = ?", n.UUID).First(n).Error
}

func (n *Node) FindAndPage(page int, pageSize int) ([]Node, int64, error) {
	db := dbclient.GetMysqlDB().Table(n.TableName())
	if len(n.UUID) > 0 {
		db = db.Where("uuid = ?", n.UUID)
	}

	if len(n.IP) > 0 {
		db = db.Where("ip = ?", n.IP)
	}

	if n.Status > 0 {
		db = db.Where("status = ?", n.Status)
	}

	if n.UpTime > 0 {
		db = db.Where("up > ?", n.UpTime)
	}

	nodes := make([]Node, 0, 2)
	var total int64
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Limit(pageSize).Offset((page-1)*pageSize).Order("up desc").Find(&nodes).Error
	if err != nil {
		return nil, 0, err
	}

	return nodes, total, nil
}

func (n *Node) GetNodeCount() (int64, error) {
	db := dbclient.GetMysqlDB().Table(n.TableName())
	if n.Status > 0 {
		db = db.Where("status = ?", n.Status)
	}

	var total int64
	err := db.Count(&total).Error
	if err != nil {
		return 0, err
	}

	return total, nil
}




