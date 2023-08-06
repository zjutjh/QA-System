package userController

import (
	"QA-system/apiException"
	"QA-system/service/configService"
	"QA-system/service/sessionServices"
	"QA-system/utility"
	"github.com/gin-gonic/gin"
)

type LoginData struct {
	Password string `json:"password"`
}

func Login(c *gin.Context) {
	var postForm LoginData
	err := c.ShouldBindJSON(&postForm)

	if err != nil {
		_ = c.AbortWithError(200, apiException.ParamError)
		return
	}
	if configService.GetConfig("password") != postForm.Password {
		_ = c.AbortWithError(200, apiException.NoThatPasswordOrWrong)
		return
	}
	err = sessionServices.SetUserSession(c)
	if err != nil {
		_ = c.AbortWithError(200, apiException.ServerError)
		return
	}
	utility.JsonSuccessResponse(c, nil)
}
