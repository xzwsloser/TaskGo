package dbclient

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/xzwsloser/TaskGo/pkg/config"
	taskGoLogger "github.com/xzwsloser/TaskGo/pkg/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var _defaultDB *gorm.DB

type writer struct {
	logger.Writer
}

func newWriter(w logger.Writer) *writer {
	return &writer{Writer: w}
}

func setMysqlConfig(logMode string) *gorm.Config {
	config := &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true}
	_default := logger.New(newWriter(log.New(os.Stdout, "\r\n", log.LstdFlags)), logger.Config{
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      logger.Warn,
		Colorful:      true,
	})
	switch logMode {
	case "silent", "Silent":
		config.Logger = _default.LogMode(logger.Silent)
	case "error", "Error":
		config.Logger = _default.LogMode(logger.Error)
	case "warn", "Warn":
		config.Logger = _default.LogMode(logger.Warn)
	case "info", "Info":
		config.Logger = _default.LogMode(logger.Info)
	default:
		config.Logger = _default.LogMode(logger.Info)
	}
	return config
}

func getDsn(mc config.Mysql) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", 
					mc.Username, mc.Password, mc.Address, mc.Port, mc.Dbname)
}

func emptyDsn(mc config.Mysql) string {
	if mc.Address == "" {
		mc.Address = "127.0.0.1"
	} 

	if mc.Port == "" {
		mc.Port = "3306"
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%s)", mc.Username, mc.Port, mc.Address, mc.Port)
}

func InitMysql() (*gorm.DB, error) {
	mc := config.GetConfig().Mysql
	dsn := getDsn(mc)

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: dsn,
		DefaultStringSize: 256,
		SkipInitializeWithVersion: false,
	}), setMysqlConfig(mc.LogMode))

	if err != nil {
		return nil, err
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(mc.MaxIdleConns)
	sqlDB.SetMaxOpenConns(mc.MaxOpenConns)
	_defaultDB = db

	return db, nil
}

func GetMysqlDB() *gorm.DB {
	if _defaultDB == nil {
		taskGoLogger.GetLogger().Error("Mysql Client Not Init.")
		return nil
	}

	return _defaultDB
}

func CreateDatabase(driver, createSql string) error {
	mc := config.GetConfig().Mysql
	dsn := emptyDsn(mc)
	db, err := sql.Open(driver, dsn)
	if err != nil {
		fmt.Printf("Failed to connect to database: %s", err)
		return err
	}

	defer func(d *sql.DB) {
		err := d.Close()
		if err != nil {
			fmt.Printf("Failed to close database: %s", err)
		}
	}(db)

	err = db.Ping()
	if err != nil {
		fmt.Printf("Failed to ping database: %s", err)
		return err
	}

	_, err = db.Exec(createSql)
	return err
}


