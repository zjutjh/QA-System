package admin

import (
	"errors"
	"fmt"

	"QA-System/internal/pkg/code"
	"QA-System/internal/pkg/utils"
	"QA-System/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type createPermissionData struct {
	UserName string `json:"username" binding:"required"`
	SurveyID string `json:"survey_id" binding:"required"`
}

// CreatePermission 创建权限
func CreatePermission(c *gin.Context) {
	var data createPermissionData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	// 鉴权
	admin, err := service.GetUserSession(c)
	if err != nil {
		code.AbortWithException(c, code.NotLogin, err)
		return
	}
	if admin.AdminType != 2 {
		code.AbortWithException(c, code.NoPermission, errors.New(admin.Username+"没有权限"))
		return
	}
	user, err := service.GetUserByName(data.UserName)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	survey, err := service.GetSurveyByUUID(data.SurveyID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	if survey.UserID == user.ID {
		code.AbortWithException(c, code.PermissionBelong, errors.New("不能给问卷所有者添加权限"))
		return
	}
	err = service.CheckPermission(user.ID, data.SurveyID)
	if err == nil {
		code.AbortWithException(c, code.PermissionExist,
			fmt.Errorf("用户%d已有问卷%s权限", user.ID, data.SurveyID))
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	// 创建权限
	err = service.CreatePermission(user.ID, data.SurveyID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	utils.JsonSuccessResponse(c, nil)
}

type deletePermissionData struct {
	UserName string `form:"username" binding:"required"`
	SurveyID string `form:"survey_id" binding:"required"`
}

// DeletePermission 删除权限
func DeletePermission(c *gin.Context) {
	var data deletePermissionData
	err := c.ShouldBindQuery(&data)
	if err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	// 鉴权
	admin, err := service.GetUserSession(c)
	if err != nil {
		code.AbortWithException(c, code.NotLogin, err)
		return
	}
	if admin.AdminType != 2 {
		code.AbortWithException(c, code.NoPermission, errors.New(admin.Username+"没有权限"))
		return
	}
	user, err := service.GetUserByName(data.UserName)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	survey, err := service.GetSurveyByUUID(data.SurveyID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	if survey.UserID == user.ID {
		code.AbortWithException(c, code.PermissionBelong, errors.New("不能删除问卷所有者的权限"))
		return
	}
	// 查询权限
	err = service.CheckPermission(user.ID, data.SurveyID)
	if err != nil {
		code.AbortWithException(c, code.PermissionNotExist, errors.New(user.Username+"权限不存在"))
		return
	}
	// 删除权限
	err = service.DeletePermission(user.ID, data.SurveyID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	utils.JsonSuccessResponse(c, nil)
}
