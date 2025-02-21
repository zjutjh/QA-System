package main

import (
	global "QA-System/internal/global/config"
	"QA-System/internal/middleware"
	"QA-System/internal/pkg/database/mongodb"
	"QA-System/internal/pkg/database/mysql"
	"QA-System/internal/pkg/log"
	_ "QA-System/internal/pkg/redis"
	"QA-System/internal/pkg/session"
	"QA-System/internal/pkg/utils"
	"QA-System/internal/router"
	"QA-System/internal/service"
	"QA-System/pkg/extension"
	_ "QA-System/plugins"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 如果配置文件中开启了调试模式
	if !global.Config.GetBool("server.debug") {
		gin.SetMode(gin.ReleaseMode)
	}
	// 初始化日志系统
	log.ZapInit()

	// 把参数给到插件，基本已弃用
	// params := map[string]any{}

	// err := extension.ExecutePlugins()
	// if err != nil {
	// 	zap.L().Error("Error executing plugins", zap.Error(err), zap.Any("params", params))
	// 	return
	// }
	manager := extension.GetDefaultManager()
	manager.LoadPlugins()
	err := manager.ExecutePlugins()
	if err != nil {
		zap.L().Error("Error executing plugins", zap.Error(err))
	}

	// 初始化数据库 和 dao
	db := mysql.Init()
	mdb := mongodb.Init()
	service.Init(db, mdb)
	if err := utils.Init(); err != nil {
		zap.L().Fatal(err.Error())
	}

	// 初始化gin
	r := gin.Default()
	r.Use(middleware.ErrHandler())
	r.NoMethod(middleware.HandleNotFound)
	r.NoRoute(middleware.HandleNotFound)
	r.Static("public/static", "./public/static")
	r.Static("public/xlsx", "./public/xlsx")
	session.Init(r)
	router.Init(r)
	err = r.Run(":" + global.Config.GetString("server.port"))
	if err != nil {
		zap.L().Fatal("Failed to start the server:" + err.Error())
	}
}
