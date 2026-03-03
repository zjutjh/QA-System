// Package bootstrap 负责应用各子模块的初始化，将 main.go 中的多职责启动流程模块化。
// 每个 Init* 函数具备单一职责，初始化错误直接返回给调用方统一处理。
package bootstrap

import (
	"time"

	global "QA-System/internal/global/config"
	"QA-System/internal/pkg/database/mongodb"
	"QA-System/internal/pkg/database/mysql"
	"QA-System/internal/pkg/idgen"
	"QA-System/internal/pkg/log"
	"QA-System/internal/pkg/oss"
	r "QA-System/internal/pkg/redis"
	"QA-System/internal/pkg/session"
	"QA-System/internal/pkg/utils"
	"QA-System/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// InitTimezone 设置应用时区为 Asia/Shanghai，失败则回退到固定偏移。
func InitTimezone() {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		zap.L().Error("Failed to load location, using fixed zone instead", zap.Error(err))
		loc = time.FixedZone("CST", 8*60*60)
	}
	time.Local = loc
}

// InitGinMode 根据配置文件设置 Gin 运行模式。
func InitGinMode() {
	if !global.Config.GetBool("server.debug") {
		gin.SetMode(gin.ReleaseMode)
	}
}

// InitLogger 初始化 zap 日志系统。
func InitLogger() {
	log.ZapInit()
}

// InitIDGenerator 初始化雪花 ID 生成器。
func InitIDGenerator() {
	idgen.Init()
}

// InitRedis 显式初始化 Redis 客户端。
func InitRedis() {
	r.Init()
}

// InitOSS 显式初始化对象存储客户端。
func InitOSS() {
	oss.Init()
}

// InitDatabase 初始化 MySQL 与 MongoDB，并注入到 service 层。
func InitDatabase() {
	db := mysql.Init()
	mdb := mongodb.Init()
	service.Init(db, mdb)
}

// InitCrypto 初始化 AES 加密工具。
func InitCrypto() error {
	return utils.Init()
}

// InitSession 初始化 Session 会话管理。
func InitSession(r *gin.Engine) {
	session.Init(r)
}
