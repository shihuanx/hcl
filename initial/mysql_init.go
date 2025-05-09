package initial

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

var DB *gorm.DB

// InitMysql 初始化数据库连接
func InitMysql(dsn string) (*gorm.DB, error) {

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	// 连接池配置
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)               // 最大空闲连接
	sqlDB.SetMaxOpenConns(100)              // 最大打开连接
	sqlDB.SetConnMaxLifetime(time.Hour * 1) // 连接最大存活时间

	DB = db
	return db, nil
}
