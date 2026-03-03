package user

import (
	"errors"
	"time"

	"QA-System/internal/pkg/code"
	"QA-System/internal/pkg/utils"
	"QA-System/internal/service"

	"github.com/gin-gonic/gin"
)

type getSurveyData struct {
	ID           int64 `form:"id" binding:"required"`
	IsPreVisible bool  `form:"is_pre_visible"`
}

// GetSurvey 用户获取问卷
func GetSurvey(c *gin.Context) {
	var data getSurveyData
	err := c.ShouldBindQuery(&data)
	if err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	// 获取用户信息
	user, err := service.GetUserSession(c)
	// Treat an empty error message as "no session" (anonymous access), but
	// still abort on any non-nil, non-empty error.
	if err != nil && err.Error() != "" {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	// 获取问卷
	survey, err := service.GetSurveyByID(data.ID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	// 判断填写时间是否在问卷有效期内
	if !survey.Deadline.IsZero() && survey.Deadline.Before(time.Now()) {
		code.AbortWithException(c, code.TimeBeyondError, errors.New("问卷填写时间已截止"))
		return
	}
	// 判断问卷是否开放
	if survey.Status != 2 && (user == nil || !data.IsPreVisible) {
		code.AbortWithException(c, code.SurveyNotOpen, errors.New("问卷未开放"))
		return
	}
	if !survey.StartTime.IsZero() && survey.StartTime.After(time.Now()) {
		code.AbortWithException(c, code.SurveyNotOpen, errors.New("问卷未开放"))
		return
	}
	// 获取相应的问题
	questions, err := service.GetQuestionsBySurveyID(survey.ID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	// 构建问卷响应
	questionListsResponse := make([]map[string]any, 0)
	for _, question := range questions {
		options, err := service.GetOptionsByQuestionID(question.ID)
		if err != nil {
			code.AbortWithException(c, code.ServerError, err)
			return
		}
		optionsResponse := make([]map[string]any, 0)
		for _, option := range options {
			optionResponse := map[string]any{
				"img":         option.Img,
				"content":     option.Content,
				"description": option.Description,
				"serial_num":  option.SerialNum,
			}
			optionsResponse = append(optionsResponse, optionResponse)
		}

		questionSettingResponse := map[string]any{
			"required":       question.Required,
			"unique":         question.Unique,
			"other_option":   question.OtherOption,
			"question_type":  question.QuestionType,
			"reg":            question.Reg,
			"maximum_option": question.MaximumOption,
			"minimum_option": question.MinimumOption,
		}

		questionListMap := map[string]any{
			"id":           question.ID,
			"serial_num":   question.SerialNum,
			"subject":      question.Subject,
			"description":  question.Description,
			"img":          question.Img,
			"ques_setting": questionSettingResponse,
			"options":      optionsResponse,
		}
		questionListsResponse = append(questionListsResponse, questionListMap)
	}

	questionsConfigResponse := map[string]any{
		"title":         survey.Title,
		"desc":          survey.Desc,
		"question_list": questionListsResponse,
	}
	baseConfigResponse := map[string]any{
		"start_time":     survey.StartTime,
		"end_time":       survey.Deadline,
		"day_limit":      survey.DailyLimit,
		"sum_limit":      survey.SumLimit,
		"verify":         survey.Verify,
		"undergrad_only": survey.UndergradOnly,
		"need_notify":    survey.NeedNotify,
	}
	response := map[string]any{
		"id":          survey.ID,
		"status":      survey.Status,
		"survey_type": survey.Type,
		"base_config": baseConfigResponse,
		"ques_config": questionsConfigResponse,
	}

	utils.JsonSuccessResponse(c, response)
}
