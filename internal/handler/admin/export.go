package admin

import (
	"errors"

	"QA-System/internal/pkg/code"
	"QA-System/internal/pkg/utils"
	"QA-System/internal/service"

	"github.com/gin-gonic/gin"
)

type downloadFileData struct {
	ID int64 `form:"id" binding:"required"`
}

// DownloadFile 下载
func DownloadFile(c *gin.Context) {
	var data downloadFileData
	err := c.ShouldBindQuery(&data)
	if err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	user, err := service.GetUserSession(c)
	if err != nil {
		code.AbortWithException(c, code.NotLogin, err)
		return
	}
	// 获取问卷
	survey, err := service.GetSurveyByID(data.ID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	// 判断权限
	if (user.AdminType != 2) && (user.AdminType != 1 || survey.UserID != user.ID) &&
		!service.UserInManage(user.ID, survey.ID) {
		code.AbortWithException(c, code.NoPermission, errors.New(user.Username+"无权限"))
		return
	}
	// 获取数据
	answers, err := service.GetAllSurveyAnswers(data.ID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	url, err := service.HandleDownloadFile(answers, survey)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	utils.JsonSuccessResponse(c, url)
}

type downloadChooseData struct {
	ID int64 `form:"id" binding:"required"`
}

// DownloadChooseFile 下载选择题数据
func DownloadChooseFile(c *gin.Context) {
	var data downloadChooseData
	err := c.ShouldBindQuery(&data)
	if err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	user, err := service.GetUserSession(c)
	if err != nil {
		code.AbortWithException(c, code.NotLogin, err)
		return
	}
	// 获取问卷
	survey, err := service.GetSurveyByID(data.ID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	// 判断权限
	if (user.AdminType != 2) && (user.AdminType != 1 || survey.UserID != user.ID) &&
		!service.UserInManage(user.ID, survey.ID) {
		code.AbortWithException(c, code.NoPermission, errors.New(user.Username+"无权限"))
		return
	}
	// 获取数据
	answers, err := service.GetSurveyAnswersBySurveyID(data.ID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	questions, err := service.GetQuestionsBySurveyID(data.ID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	stats := service.GenerateQuestionStats(questions, answers)
	url, err := service.HandleChooseStatistics(survey, stats)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	utils.JsonSuccessResponse(c, url)
}
