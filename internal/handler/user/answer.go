package user

import (
	"errors"

	"QA-System/internal/pkg/code"
	"QA-System/internal/pkg/utils"
	"QA-System/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/WeJH-SDK/oauth"
)

type getAnswerRecordData struct {
	SurveyID int64  `json:"survey_id" form:"survey_id" binding:"required"`
	Token    string `json:"token" form:"token" binding:"required"`
}

// GetAnswerRecord 用户获取问卷填写记录
func GetAnswerRecord(c *gin.Context) {
	var data getAnswerRecordData
	err := c.ShouldBindQuery(&data)
	if err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	survey, err := service.GetSurveyByID(data.SurveyID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	if !survey.Verify {
		code.AbortWithException(c, code.SurveyTypeError, errors.New("问卷非需统一验证问卷"))
		return
	}
	// 获取用户信息
	var userInfo oauth.UserInfo
	userInfo, err = utils.ParseJWT(data.Token)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}

	// 获取该用户对于该问卷的所有填写记录
	userAnswerSheets, err := service.GetSurveyAnswersBySurveyIDAndStudentID(data.SurveyID, userInfo.StudentID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}

	// 获取问题信息
	questions, err := service.GetQuestionsBySurveyID(data.SurveyID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}

	records := service.CreateRecordDetailResponseWithQuestions(userAnswerSheets, questions)

	utils.JsonSuccessResponse(c, gin.H{
		"records": records,
	})
}
