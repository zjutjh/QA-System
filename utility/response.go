package utility

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func JsonResponse(code int, msg string, data gin.H, c *gin.Context) {
	c.JSON(200, gin.H{
		"code": code,
		"msg":  msg,
		"data": data,
	})
}

func JsonSuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"data": data,
		"code": 200,
		"msg":  "ok",
	})
}
