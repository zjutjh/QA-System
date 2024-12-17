package session

import (
	"github.com/gin-gonic/gin"
	WeJHSDK "github.com/zjutjh/WeJH-SDK"
)

func Init(r *gin.Engine) {
	config := getConfig()
	WeJHSDK.SessionInit(r, config)
}
