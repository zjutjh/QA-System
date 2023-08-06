package midwares

import (
	"QA-system/apiException"
	"QA-system/service/sessionServices"
	"github.com/gin-gonic/gin"
)

func CheckLogin(c *gin.Context) {
	isLogin := sessionServices.CheckUserSession(c)
	if !isLogin {
		_ = c.AbortWithError(200, apiException.NotLogin)
		return
	}
	c.Next()
}
