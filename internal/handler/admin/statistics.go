package admin

import (
	"errors"
	"math"

	"QA-System/internal/pkg/code"
	"QA-System/internal/pkg/utils"
	"QA-System/internal/service"

	"github.com/gin-gonic/gin"
)

type getSurveyStatisticsData struct {
	ID       int64 `form:"id" binding:"required"`
	PageNum  int   `form:"page_num" binding:"required,gt=0"`
	PageSize int   `form:"page_size" binding:"required,gt=0"`
}

// GetSurveyStatistics 获取统计问卷选择题数据
func GetSurveyStatistics(c *gin.Context) {
	var data getSurveyStatisticsData
	if err := c.ShouldBindQuery(&data); err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}

	user, err := service.GetUserSession(c)
	if err != nil {
		code.AbortWithException(c, code.NotLogin, err)
		return
	}

	survey, err := service.GetSurveyByID(data.ID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}

	if (user.AdminType != 2) && (user.AdminType != 1 || survey.UserID != user.ID) &&
		!service.UserInManage(user.ID, survey.ID) {
		code.AbortWithException(c, code.NoPermission, errors.New(user.Username+"无权限"))
		return
	}

	answersheets, err := service.GetSurveyAnswersBySurveyID(data.ID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}

	questions, err := service.GetQuestionsBySurveyID(data.ID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}

	response := service.GenerateQuestionStats(questions, answersheets)
	start := (data.PageNum - 1) * data.PageSize
	end := start + data.PageSize
	// 确保 start 和 end 在有效范围内
	if start < 0 {
		start = 0
	}
	if end > len(response) {
		end = len(response)
	}
	if start > end {
		start = end
	}

	// 访问切片
	resp := response[start:end]
	totalSumPage := math.Ceil(float64(len(response)) / float64(data.PageSize))

	utils.JsonSuccessResponse(c, gin.H{
		"statistics":     resp,
		"total":          len(answersheets),
		"total_sum_page": totalSumPage,
		"survey_type":    survey.Type,
	})
}
