package admin

import (
	"errors"

	"QA-System/internal/pkg/code"
	"QA-System/internal/pkg/utils"
	"QA-System/internal/service"

	"github.com/gin-gonic/gin"
)

type getQuestionPreData struct {
	Type string `form:"type"`
}

// GetQuestionPre 获取预先信息
func GetQuestionPre(c *gin.Context) {
	var data getQuestionPreData
	if err := c.ShouldBindQuery(&data); err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}

	user, err := service.GetUserSession(c)
	if err != nil {
		code.AbortWithException(c, code.NotLogin, err)
		return
	}

	if (user.AdminType != 2) && (user.AdminType != 1) {
		code.AbortWithException(c, code.NoPermission, errors.New(user.Username+"无权限"))
		return
	}

	// 获取预先信息
	value, err := service.GetQuestionPre(data.Type)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	utils.JsonSuccessResponse(c, gin.H{
		"value": value,
	})
}

type createQuestionPreData struct {
	Type  string   `json:"type"`
	Value []string `json:"value"`
}

// CreateQuestionPre 创建预先信息
func CreateQuestionPre(c *gin.Context) {
	var data createQuestionPreData
	if err := c.ShouldBindJSON(&data); err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}

	user, err := service.GetUserSession(c)
	if err != nil {
		code.AbortWithException(c, code.NotLogin, err)
		return
	}

	if (user.AdminType != 2) && (user.AdminType != 1) {
		code.AbortWithException(c, code.NoPermission, errors.New(user.Username+"无权限"))
		return
	}

	// 创建预先信息
	err = service.CreateQuestionPre(data.Type, data.Value)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	utils.JsonSuccessResponse(c, nil)
}
