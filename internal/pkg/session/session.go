package session

import (
	"github.com/gin-gonic/gin"
	WeJHSDK "github.com/zjutjh/WeJH-SDK"
)

// Init 初始化Session会话管理
func Init(r *gin.Engine) {
	config := getConfig()
	WeJHSDK.SessionInit(r, config)
}
