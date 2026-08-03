package db

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"hunjie/internal/config"
)

// DB 全局数据库句柄。service 直接用它写查询，不做 repository 层
var DB *gorm.DB

func Init() error {
	logLevel := logger.Info
	if config.C.Server.Mode == "release" {
		logLevel = logger.Warn
	}

	d, err := gorm.Open(mysql.Open(config.C.MySQL.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // User → user，与建表脚本一致
		},
	})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlDB, err := d.DB()
	if err != nil {
		return fmt.Errorf("获取底层连接失败: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库 ping 失败: %w", err)
	}

	DB = d
	return nil
}
