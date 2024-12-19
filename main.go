package main

import (
	global "QA-System/internal/global/config"
	"QA-System/internal/middleware"
	mongodb "QA-System/internal/pkg/database/mongodb"
	mysql "QA-System/internal/pkg/database/mysql"
	"QA-System/internal/pkg/log"
	"QA-System/internal/pkg/queue/asynq"
	"QA-System/internal/pkg/session"
	"QA-System/internal/router"
	"QA-System/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// 如果配置文件中开启了调试模式
	if !global.Config.GetBool("server.debug") {
		gin.SetMode(gin.ReleaseMode)
	}
	// 初始化日志系统
	log.ZapInit()
	// 初始化数据库
	db := mysql.MysqlInit()
	mdb := mongodb.MongodbInit()
	// 初始化dao
	service.ServiceInit(db, mdb)

	// 初始化gin
	r := gin.Default()
	r.Use(middlewares.ErrHandler())
	r.NoMethod(middlewares.HandleNotFound)
	r.NoRoute(middlewares.HandleNotFound)
	r.Static("public/static", "./public/static")
	r.Static("public/xlsx", "./public/xlsx")
	session.Init(r)
	router.Init(r)
	go asynq.AsynqInit()
	err := r.Run(":" + global.Config.GetString("server.port"))
	if err != nil {
		log.Logger.Fatal("Failed to start the server:" + err.Error())
	}
}
