package admin

import (
	"QA-System/internal/pkg/code"
	"QA-System/internal/pkg/utils"
	"QA-System/internal/service"
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreatePermissionData struct {
	UserName string `json:"username" binding:"required"`
	SurveyID int    `json:"survey_id" binding:"required"`
}

func CreatrPermission(c *gin.Context) {
	var data CreatePermissionData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取参数失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ParamError)
		return
	}
	//鉴权
	admin, err := service.GetUserSession(c)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取用户缓存信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NotLogin)
		return
	}
	if admin.AdminType != 2 {
		c.Error(&gin.Error{Err: errors.New(admin.Username + "没有权限"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NoPermission)
		return
	}
	user, err := service.GetUserByName(data.UserName)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取用户信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	survey, err := service.GetSurveyByID(data.SurveyID)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取问卷信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	if survey.UserID == user.ID {
		c.Error(&gin.Error{Err: errors.New("不能给问卷所有者添加权限"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.PermissionBelong)
		return
	}
	err = service.CheckPermission(user.ID, data.SurveyID)
	if err == nil {
		c.Error(&gin.Error{Err: errors.New("权限已存在"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.PermissionExist)
		return
	} else if err != gorm.ErrRecordNotFound {
		c.Error(&gin.Error{Err: errors.New("查询权限失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	//创建权限
	err = service.CreatePermission(user.ID, data.SurveyID)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("创建权限失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	utils.JsonSuccessResponse(c, nil)
}

type DeletePermissionData struct {
	UserName string `form:"username" binding:"required"`
	SurveyID int    `form:"survey_id" binding:"required"`
}

func DeletePermission(c *gin.Context) {
	var data DeletePermissionData
	err := c.ShouldBindQuery(&data)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取参数失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ParamError)
		return
	}
	//鉴权
	admin, err := service.GetUserSession(c)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取用户缓存信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NotLogin)
		return
	}
	if admin.AdminType != 2 {
		c.Error(&gin.Error{Err: errors.New(admin.Username + "没有权限"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NoPermission)
		return
	}
	user, err := service.GetUserByName(data.UserName)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取用户信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	survey, err := service.GetSurveyByID(data.SurveyID)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取问卷信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	if survey.UserID == user.ID {
		c.Error(&gin.Error{Err: errors.New("不能删除问卷所有者的权限"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.PermissionBelong)
		return
	}
	//查询权限
	err = service.CheckPermission(user.ID, data.SurveyID)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New(user.Username + "权限不存在"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.PermissionNotExist)
		return
	}
	//删除权限
	err = service.DeletePermission(user.ID, data.SurveyID)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("删除权限失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	utils.JsonSuccessResponse(c, nil)
}
