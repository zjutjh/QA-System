package main

import (
	"QA-System/internal/bootstrap"
	global "QA-System/internal/global/config"
	_ "QA-System/plugins"

	"go.uber.org/zap"
)

func main() {
	// ── 基础设施初始化 ──
	bootstrap.InitTimezone()
	bootstrap.InitGinMode()
	bootstrap.InitLogger()
	bootstrap.InitIDGenerator()
	bootstrap.InitRedis()
	bootstrap.InitOSS()
	bootstrap.InitDatabase()
	if err := bootstrap.InitCrypto(); err != nil {
		zap.L().Fatal(err.Error())
	}

	// ── 插件加载 ──
	bootstrap.InitPlugins()

	// ── HTTP 服务启动 ──
	r := bootstrap.NewHTTPEngine()
	if err := r.Run(":" + global.Config.GetString("server.port")); err != nil {
		zap.L().Fatal("Failed to start the server:" + err.Error())
	}
}
