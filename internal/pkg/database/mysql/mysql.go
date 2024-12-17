package database

import (
	"fmt"

	"QA-System/internal/global/config"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func MysqlInit() *gorm.DB {

	user := global.Config.GetString("mysql.user")
	pass := global.Config.GetString("mysql.pass")
	port := global.Config.GetString("mysql.port")
	host := global.Config.GetString("mysql.host")
	name := global.Config.GetString("mysql.name")

	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8&parseTime=True&loc=Local", user, pass, host, port, name)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true, // 关闭外键约束 提升数据库速度
	})

	if err != nil {
		zap.L().Fatal("Failed to connect to MySQL:" + err.Error())
	}

	err = autoMigrate(db)
	if err != nil {
		zap.L().Fatal("DatabaseMigrateFailed" + err.Error())
	}
	zap.L().Info("Connected to MySQL")
	return db
}
