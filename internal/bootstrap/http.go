package bootstrap

import (
	"QA-System/internal/middleware"
	"QA-System/internal/router"

	"github.com/gin-gonic/gin"
)

// NewHTTPEngine 创建并配置 Gin 引擎，注册中间件、静态资源、Session 和路由。
func NewHTTPEngine() *gin.Engine {
	r := gin.Default()
	r.Use(middleware.ErrHandler())
	r.NoMethod(middleware.HandleNotFound)
	r.NoRoute(middleware.HandleNotFound)
	r.Static("public/static", "./public/static")
	r.Static("public/xlsx", "./public/xlsx")
	InitSession(r)
	router.Init(r)
	return r
}
