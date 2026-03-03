package admin

import (
	"errors"
	"math"
	"strconv"
	"time"

	"QA-System/internal/model"
	"QA-System/internal/pkg/code"
	"QA-System/internal/pkg/utils"
	"QA-System/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type createSurveyData struct {
	Status         int                  `json:"status" binding:"required,oneof=1 2"`
	SurveyType     uint                 `json:"survey_type"` // 问卷类型 0:调研 1:投票
	BaseConfig     model.BaseConfig     `json:"base_config"` // 基本配置
	QuestionConfig model.QuestionConfig `json:"ques_config"` // 问题设置
}

// CreateSurvey 创建问卷
func CreateSurvey(c *gin.Context) {
	var data createSurveyData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	// 鉴权
	user, err := service.GetUserSession(c)
	if err != nil {
		code.AbortWithException(c, code.NotLogin, err)
		return
	}
	// 解析时间转换为中国时间(UTC+8)
	ddlTime, err := time.Parse(time.RFC3339, data.BaseConfig.EndTime)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	startTime, err := time.Parse(time.RFC3339, data.BaseConfig.StartTime)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	if startTime.After(ddlTime) {
		code.AbortWithException(c, code.SurveyError, errors.New("开始时间晚于截止时间"))
		return
	}
	// 检查总投票次数大于日投票数
	if data.BaseConfig.SumLimit != 0 && data.BaseConfig.DailyLimit != 0 &&
		data.BaseConfig.SumLimit < data.BaseConfig.DailyLimit {
		code.AbortWithException(c, code.SurveyError, errors.New("总投票次数小于单日投票次数"))
		return
	}
	// 检查问卷每个题目的序号没有重复且按照顺序递增
	questionNumMap := make(map[int]bool)
	for i, question := range data.QuestionConfig.QuestionList {
		if data.SurveyType == 1 && (question.QuestionSetting.QuestionType == 1 && !question.QuestionSetting.Required) {
			code.AbortWithException(c, code.SurveyError, errors.New("投票题目只能为多选必填题"))
			return
		}
		if questionNumMap[question.SerialNum] {
			code.AbortWithException(c, code.SurveyError, errors.New("题目序号"+strconv.Itoa(question.SerialNum)+"重复"))
			return
		}
		if i > 0 && question.SerialNum != data.QuestionConfig.QuestionList[i-1].SerialNum+1 {
			code.AbortWithException(c, code.SurveyError, errors.New("题目序号不按顺序递增"))
			return
		}
		questionNumMap[question.SerialNum] = true
		question.SerialNum = i + 1

		// 检测多选题目的最多选项数和最少选项数
		if ((question.QuestionSetting.QuestionType == 2 && data.SurveyType == 0) ||
			(question.QuestionSetting.QuestionType == 1 && data.SurveyType == 1)) &&
			(question.QuestionSetting.MaximumOption < question.QuestionSetting.MinimumOption) {
			code.AbortWithException(c, code.OptionNumError, errors.New("多选最多选项数小于最少选项数"))
			return
		}
		// 检查多选选项和最少选项数是否符合要求
		if ((question.QuestionSetting.QuestionType == 2 && data.SurveyType == 0) ||
			(question.QuestionSetting.QuestionType == 1 && data.SurveyType == 1)) &&
			uint(len(question.Options)) < question.QuestionSetting.MinimumOption {
			code.AbortWithException(c, code.OptionNumError, errors.New("选项数量小于最少选项数"))
			return
		}
		// 检查最多选项数是否符合要求
		if ((question.QuestionSetting.QuestionType == 2 && data.SurveyType == 0) ||
			(question.QuestionSetting.QuestionType == 1 && data.SurveyType == 1)) &&
			question.QuestionSetting.MaximumOption == 0 {
			code.AbortWithException(c, code.OptionNumError, errors.New("最多选项数小于等于0"))
			return
		}
	}
	// 检测问卷是否填写完整
	if data.Status == 2 {
		if data.QuestionConfig.Title == "" || len(data.QuestionConfig.QuestionList) == 0 {
			code.AbortWithException(c, code.SurveyIncomplete, errors.New("问卷标题为空或问卷没有问题"))
			return
		}
		questionMap := make(map[string]bool)
		for _, question := range data.QuestionConfig.QuestionList {
			if question.Subject == "" {
				code.AbortWithException(c, code.SurveyIncomplete,
					errors.New("问题"+strconv.Itoa(question.SerialNum)+"标题为空"))
				return
			}
			if questionMap[question.Subject] {
				code.AbortWithException(c, code.SurveyContentRepeat,
					errors.New("问题"+strconv.Itoa(question.SerialNum)+"题目"+question.Subject+"重复"))
				return
			}
			questionMap[question.Subject] = true
			if question.QuestionSetting.QuestionType == 1 || question.QuestionSetting.QuestionType == 2 {
				if len(question.Options) < 1 {
					code.AbortWithException(c, code.SurveyIncomplete,
						errors.New("问题"+strconv.Itoa(question.SerialNum)+"选项数量太少"))
					return
				}
				optionMap := make(map[string]bool)
				for _, option := range question.Options {
					if option.Content == "" {
						code.AbortWithException(c, code.SurveyIncomplete,
							errors.New("选项"+strconv.Itoa(option.SerialNum)+"内容为空"))
						return
					}
					if optionMap[option.Content] {
						code.AbortWithException(c, code.SurveyContentRepeat,
							errors.New("选项内容"+option.Content+"重复"))
						return
					}
					optionMap[option.Content] = true
				}
			}
		}
	}
	// 创建问卷
	err = service.CreateSurvey(user.ID, data.QuestionConfig.QuestionList, data.Status, data.SurveyType, data.BaseConfig.
		DailyLimit, data.BaseConfig.SumLimit, data.BaseConfig.Verify, data.BaseConfig.UndergradOnly, ddlTime, startTime,
		data.QuestionConfig.Title, data.QuestionConfig.Desc, data.BaseConfig.NeedNotify)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	utils.JsonSuccessResponse(c, nil)
}

type updateSurveyStatusData struct {
	ID     int64 `json:"id" binding:"required"`
	Status int   `json:"status" binding:"required,oneof=1 2"`
}

// UpdateSurveyStatus 修改问卷状态
func UpdateSurveyStatus(c *gin.Context) {
	var data updateSurveyStatusData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	// 鉴权 + 权限检查
	_, survey, ok := requireSurveyPermission(c, data.ID)
	if !ok {
		return
	}
	// 判断问卷状态
	if survey.Status == data.Status {
		code.AbortWithException(c, code.StatusRepeatError, errors.New("问卷状态重复"))
		return
	}
	// 检测问卷是否填写完整
	if data.Status == 2 {
		if survey.Title == "" {
			code.AbortWithException(c, code.SurveyIncomplete, errors.New("问卷信息填写不完整"))
			return
		}
		questions, err := service.GetQuestionsBySurveyID(survey.ID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			code.AbortWithException(c, code.SurveyIncomplete, errors.New("问卷问题不存在"))
			return
		} else if err != nil {
			code.AbortWithException(c, code.ServerError, err)
			return
		}
		questionMap := make(map[string]bool)
		for _, question := range questions {
			if question.Subject == "" {
				code.AbortWithException(c, code.SurveyIncomplete,
					errors.New("问题"+strconv.Itoa(question.SerialNum)+"内容填写为空"))
				return
			}
			if questionMap[question.Subject] {
				code.AbortWithException(c, code.SurveyContentRepeat,
					errors.New("问题题目"+question.Subject+"重复"))
				return
			}
			questionMap[question.Subject] = true
			if question.QuestionType == 1 || question.QuestionType == 2 {
				options, err := service.GetOptionsByQuestionID(question.ID)
				if err != nil {
					code.AbortWithException(c, code.ServerError, err)
					return
				}
				if len(options) < 1 {
					code.AbortWithException(c, code.SurveyIncomplete,
						errors.New("问题"+strconv.Itoa(question.ID)+"选项太少"))
					return
				}
				optionMap := make(map[string]bool)
				for _, option := range options {
					if option.Content == "" {
						code.AbortWithException(c, code.SurveyIncomplete,
							errors.New("选项"+strconv.Itoa(option.SerialNum)+"内容未填"))
						return
					}
					if optionMap[option.Content] {
						code.AbortWithException(c, code.SurveyContentRepeat,
							errors.New("选项内容"+option.Content+"重复"))
						return
					}
					optionMap[option.Content] = true
				}
			}
		}
	}
	// 修改问卷状态
	err = service.UpdateSurveyStatus(data.ID, data.Status)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	utils.JsonSuccessResponse(c, nil)
}

type updateSurveyData struct {
	ID             int64                `json:"id" binding:"required"`
	SurveyType     uint                 `json:"survey_type"` // 问卷类型 0:调研 1:投票
	BaseConfig     model.BaseConfig     `json:"base_config"` // 基本配置
	QuestionConfig model.QuestionConfig `json:"ques_config"` // 问题设置
}

// UpdateSurvey 修改问卷
func UpdateSurvey(c *gin.Context) {
	var data updateSurveyData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	// 鉴权
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
	// 判断问卷状态
	if user.AdminType != 2 {
		if survey.Status != 1 {
			code.AbortWithException(c, code.StatusOpenError, errors.New("问卷状态不为未发布"))
			return
		}
		// 判断问卷的填写数量是否为零
		if survey.Num != 0 {
			code.AbortWithException(c, code.SurveyNumError, errors.New("问卷已有填写数量"))
			return
		}
	}
	// 解析时间转换为中国时间(UTC+8)
	ddlTime, err := time.Parse(time.RFC3339, data.BaseConfig.EndTime)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	startTime, err := time.Parse(time.RFC3339, data.BaseConfig.StartTime)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	if startTime.After(ddlTime) {
		code.AbortWithException(c, code.SurveyError, errors.New("开始时间晚于截止时间"))
		return
	}
	if data.BaseConfig.SumLimit != 0 && data.BaseConfig.DailyLimit != 0 &&
		data.BaseConfig.SumLimit < data.BaseConfig.DailyLimit {
		code.AbortWithException(c, code.SurveyError, errors.New("总投票次数小于单日投票次数"))
		return
	}
	// 检查问卷每个题目的序号没有重复且按照顺序递增
	questionNumMap := make(map[int]bool)
	for i, question := range data.QuestionConfig.QuestionList {
		if questionNumMap[question.SerialNum] {
			code.AbortWithException(c, code.SurveyError, errors.New("题目序号"+strconv.Itoa(question.SerialNum)+"重复"))
			return
		}
		if i > 0 && question.SerialNum != data.QuestionConfig.QuestionList[i-1].SerialNum+1 {
			code.AbortWithException(c, code.SurveyError, errors.New("题目序号不按顺序递增"))
			return
		}
		questionNumMap[question.SerialNum] = true
		question.SerialNum = i + 1

		// 检测多选题目的最多选项数和最少选项数
		if ((question.QuestionSetting.QuestionType == 2 && data.SurveyType == 0) ||
			(question.QuestionSetting.QuestionType == 1 && data.SurveyType == 1)) &&
			(question.QuestionSetting.MaximumOption < question.QuestionSetting.MinimumOption) {
			code.AbortWithException(c, code.OptionNumError, errors.New("多选最多选项数小于最少选项数"))
			return
		}
		// 检查多选选项和最少选项数是否符合要求
		if ((question.QuestionSetting.QuestionType == 2 && data.SurveyType == 0) ||
			(question.QuestionSetting.QuestionType == 1 && data.SurveyType == 1)) &&
			uint(len(question.Options)) < question.QuestionSetting.MinimumOption {
			code.AbortWithException(c, code.OptionNumError, errors.New("选项数量小于最少选项数"))
			return
		}
		// 检查最多选项数是否符合要求
		if ((question.QuestionSetting.QuestionType == 2 && data.SurveyType == 0) ||
			(question.QuestionSetting.QuestionType == 1 && data.SurveyType == 1)) &&
			question.QuestionSetting.MaximumOption == 0 {
			code.AbortWithException(c, code.OptionNumError, errors.New("最多选项数小于等于0"))
			return
		}
	}
	// 修改问卷
	err = service.UpdateSurvey(data.ID, data.QuestionConfig.QuestionList, data.SurveyType, data.BaseConfig.DailyLimit,
		data.BaseConfig.SumLimit, data.BaseConfig.Verify, data.BaseConfig.UndergradOnly, data.QuestionConfig.Desc,
		data.QuestionConfig.Title, ddlTime, startTime, data.BaseConfig.NeedNotify)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	utils.JsonSuccessResponse(c, nil)
}

type deleteSurveyData struct {
	ID int64 `form:"id" binding:"required"`
}

// DeleteSurvey 删除问卷
func DeleteSurvey(c *gin.Context) {
	var data deleteSurveyData
	err := c.ShouldBindQuery(&data)
	if err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	// 鉴权
	user, err := service.GetUserSession(c)
	if err != nil {
		code.AbortWithException(c, code.NotLogin, err)
		return
	}
	// 获取问卷
	survey, err := service.GetSurveyByID(data.ID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		code.AbortWithException(c, code.SurveyNotExist, errors.New("问卷不存在"))
		return
	} else if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	// 判断权限
	if (user.AdminType != 2) && (user.AdminType != 1 || survey.UserID != user.ID) &&
		!service.UserInManage(user.ID, survey.ID) {
		code.AbortWithException(c, code.NoPermission, errors.New(user.Username+"无权限"))
		return
	}
	// 删除问卷
	err = service.DeleteSurvey(data.ID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	err = service.DeleteOauthRecord(data.ID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	utils.JsonSuccessResponse(c, nil)
}

type getAllSurveyData struct {
	PageNum  int    `form:"page_num" binding:"required,gt=0"`
	PageSize int    `form:"page_size" binding:"required,gt=0"`
	Title    string `form:"title"`
}

// GetAllSurvey 获取所有问卷
func GetAllSurvey(c *gin.Context) {
	var data getAllSurveyData
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
	var surveys []model.Survey
	if user.AdminType == 2 {
		surveys, err = service.GetAllSurvey()
		if err != nil {
			code.AbortWithException(c, code.ServerError, err)
			return
		}
	} else {
		surveys, err = service.GetSurveyByUserID(user.ID)
		if err != nil {
			code.AbortWithException(c, code.ServerError, err)
			return
		}
		managedSurveys, err := service.GetManagedSurveyByUserID(user.ID)
		if err != nil {
			code.AbortWithException(c, code.ServerError, err)
			return
		}
		for _, manage := range managedSurveys {
			managedSurvey, err := service.GetSurveyByID(manage.SurveyID)
			if err != nil {
				code.AbortWithException(c, code.ServerError, err)
				return
			}
			surveys = append(surveys, *managedSurvey)
		}
	}
	surveys = service.SortSurvey(surveys)
	response := service.GetSurveyResponse(surveys)
	response, totalPageNum := service.ProcessResponse(response, data.PageNum, data.PageSize, data.Title)

	utils.JsonSuccessResponse(c, gin.H{
		"survey_list":    response,
		"total_page_num": math.Ceil(float64(totalPageNum) / float64(data.PageSize)),
	})
}

type getSurveyData struct {
	ID int64 `form:"id" binding:"required"`
}

// GetSurvey 管理员获取问卷题面
func GetSurvey(c *gin.Context) {
	var data getSurveyData
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
				"id":          option.ID,
				"serial_num":  option.SerialNum,
				"content":     option.Content,
				"img":         option.Img,
				"description": option.Description,
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
