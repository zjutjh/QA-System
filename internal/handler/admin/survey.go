package admin

import (
	"QA-System/internal/dao"
	"QA-System/internal/models"
	"QA-System/internal/pkg/code"
	"QA-System/internal/pkg/utils"
	"QA-System/internal/service"
	"errors"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

// 新建问卷
type CreateSurveyData struct {
	Title     string         `json:"title"`
	Desc      string         `json:"desc" `
	Img       string         `json:"img" `
	Status    int            `json:"status" binding:"required,oneof=1 2"`
	Time      string         `json:"time"`
	Questions []dao.Question `json:"questions"`
}

func CreateSurvey(c *gin.Context) {
	var data CreateSurveyData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取参数失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ParamError)
		return
	}
	//鉴权
	user, err := service.GetUserSession(c)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取用户缓存信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NotLogin)
		return
	}
	//解析时间转换为中国时间(UTC+8)
	ddlTime, err := time.Parse(time.RFC3339, data.Time)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("时间解析失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	// 检查问卷每个题目的序号没有重复且按照顺序递增
	questionNumMap := make(map[int]bool)
	for i, question := range data.Questions {
		if questionNumMap[question.SerialNum] {
			c.Error(&gin.Error{Err: errors.New("题目序号" + strconv.Itoa(question.SerialNum) + "重复"), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.ServerError)
			return
		}
		if i > 0 && question.SerialNum != data.Questions[i-1].SerialNum+1 {
			c.Error(&gin.Error{Err: errors.New("题目序号不按顺序递增"), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.ServerError)
			return
		}
		questionNumMap[question.SerialNum] = true
		question.SerialNum = i + 1
	}
	//检测问卷是否填写完整
	if data.Status == 2 {
		if data.Title == "" || len(data.Questions) == 0 {
			c.Error(&gin.Error{Err: errors.New("问卷标题为空或问卷没有问题"), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.SurveyIncomplete)
			return
		}
		questionMap := make(map[string]bool)
		for _, question := range data.Questions {
			if question.Subject == "" {
				c.Error(&gin.Error{Err: errors.New("问题" + strconv.Itoa(question.SerialNum) + "标题为空"), Type: gin.ErrorTypeAny})
				utils.JsonErrorResponse(c, code.SurveyIncomplete)
				return
			}
			if questionMap[question.Subject] {
				c.Error(&gin.Error{Err: errors.New("问题" + strconv.Itoa(question.SerialNum) + "题目" + question.Subject + "重复"), Type: gin.ErrorTypeAny})
				utils.JsonErrorResponse(c, code.SurveyContentRepeat)
				return
			}
			questionMap[question.Subject] = true
			if question.QuestionType == 1 || question.QuestionType == 2 {
				if len(question.Options) < 1 {
					c.Error(&gin.Error{Err: errors.New("问题" + strconv.Itoa(question.SerialNum) + "选项数量太少"), Type: gin.ErrorTypeAny})
					utils.JsonErrorResponse(c, code.SurveyIncomplete)
					return
				}
				optionMap := make(map[string]bool)
				for _, option := range question.Options {
					if option.Content == "" {
						c.Error(&gin.Error{Err: errors.New("选项" + strconv.Itoa(option.SerialNum) + "内容为空"), Type: gin.ErrorTypeAny})
						utils.JsonErrorResponse(c, code.SurveyIncomplete)
						return
					}
					if optionMap[option.Content] {
						c.Error(&gin.Error{Err: errors.New("选项内容" + option.Content + "重复"), Type: gin.ErrorTypeAny})
						utils.JsonErrorResponse(c, code.SurveyContentRepeat)
						return
					}
					optionMap[option.Content] = true
				}
			}
		}
	}
	//创建问卷
	err = service.CreateSurvey(user.ID, data.Title, data.Desc, data.Img, data.Questions, data.Status, ddlTime)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("创建问卷失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	utils.JsonSuccessResponse(c, nil)
}

// 修改问卷状态
type UpdateSurveyStatusData struct {
	ID     int `json:"id" binding:"required"`
	Status int `json:"status" binding:"required,oneof=1 2"`
}

func UpdateSurveyStatus(c *gin.Context) {
	var data UpdateSurveyStatusData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取参数失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ParamError)
		return
	}
	//鉴权
	user, err := service.GetUserSession(c)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取用户缓存信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NotLogin)
		return
	}
	// 获取问卷
	survey, err := service.GetSurveyByID(data.ID)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取问卷信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	//判断权限
	if (user.AdminType != 2) && (user.AdminType != 1 || survey.UserID != user.ID) && !service.UserInManage(user.ID, survey.ID) {
		c.Error(&gin.Error{Err: errors.New(user.Username + "无权限"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NoPermission)
		return
	}
	//判断问卷状态
	if survey.Status == data.Status {
		c.Error(&gin.Error{Err: errors.New("问卷状态重复"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.StatusRepeatError)
		return
	}
	//检测问卷是否填写完整
	if data.Status == 2 {
		if survey.Title == "" {
			c.Error(&gin.Error{Err: errors.New("问卷信息填写不完整"), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.SurveyIncomplete)
			return
		}
		questions, err := service.GetQuestionsBySurveyID(survey.ID)
		if err == gorm.ErrRecordNotFound {
			c.Error(&gin.Error{Err: errors.New("问卷问题不存在"), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.SurveyIncomplete)
			return
		} else if err != nil {
			c.Error(&gin.Error{Err: errors.New("获取问题失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.ServerError)
			return
		}
		questionMap := make(map[string]bool)
		for _, question := range questions {
			if question.Subject == "" {
				c.Error(&gin.Error{Err: errors.New("问题" + strconv.Itoa(question.SerialNum) + "内容填写为空"), Type: gin.ErrorTypeAny})
				utils.JsonErrorResponse(c, code.SurveyIncomplete)
				return
			}
			if questionMap[question.Subject] {
				c.Error(&gin.Error{Err: errors.New("问题题目" + question.Subject + "重复"), Type: gin.ErrorTypeAny})
				utils.JsonErrorResponse(c, code.SurveyContentRepeat)
				return
			}
			questionMap[question.Subject] = true
			if question.QuestionType == 1 || question.QuestionType == 2 {
				options, err := service.GetOptionsByQuestionID(question.ID)
				if err != nil {
					c.Error(&gin.Error{Err: errors.New("获取选项失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
					utils.JsonErrorResponse(c, code.ServerError)
					return
				}
				if len(options) < 1 {
					c.Error(&gin.Error{Err: errors.New("问题" + strconv.Itoa(question.ID) + "选项太少"), Type: gin.ErrorTypeAny})
					utils.JsonErrorResponse(c, code.SurveyIncomplete)
					return
				}
				optionMap := make(map[string]bool)
				for _, option := range options {
					if option.Content == "" {
						c.Error(&gin.Error{Err: errors.New("选项" + strconv.Itoa(option.SerialNum) + "内容未填"), Type: gin.ErrorTypeAny})
						utils.JsonErrorResponse(c, code.SurveyIncomplete)
						return
					}
					if optionMap[option.Content] {
						c.Error(&gin.Error{Err: errors.New("选项内容" + option.Content + "重复"), Type: gin.ErrorTypeAny})
						utils.JsonErrorResponse(c, code.SurveyContentRepeat)
						return
					}
					optionMap[option.Content] = true
				}
			}
		}
	}
	//修改问卷状态
	err = service.UpdateSurveyStatus(data.ID, data.Status)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("修改问卷状态失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	utils.JsonSuccessResponse(c, nil)
}

type UpdateSurveyData struct {
	ID        int            `json:"id" binding:"required"`
	Title     string         `json:"title"`
	Desc      string         `json:"desc" `
	Img       string         `json:"img" `
	Time      string         `json:"time"`
	Questions []dao.Question `json:"questions"`
}

func UpdateSurvey(c *gin.Context) {
	var data UpdateSurveyData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取参数失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ParamError)
		return
	}
	//鉴权
	user, err := service.GetUserSession(c)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取用户缓存信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NotLogin)
		return
	}
	// 获取问卷
	survey, err := service.GetSurveyByID(data.ID)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取问卷信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	//判断权限
	if (user.AdminType != 2) && (user.AdminType != 1 || survey.UserID != user.ID) && !service.UserInManage(user.ID, survey.ID) {
		c.Error(&gin.Error{Err: errors.New(user.Username + "无权限"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NoPermission)
		return
	}
	//判断用户权限
	if user.AdminType == 2 {
		// 	超管可以直接修改问卷
	} else if user.AdminType == 1 && survey.UserID == user.ID {
		//判断问卷状态
		if survey.Status != 1 {
			c.Error(&gin.Error{Err: errors.New("问卷状态不为未发布"), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.StatusRepeatError)
			return
		}
		// 判断问卷的填写数量是否为零
		if survey.Num != 0 {
			c.Error(&gin.Error{Err: errors.New("问卷已有填写数量"), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.SurveyNumError)
			return
		}
	} else {
		c.Error(&gin.Error{Err: errors.New(user.Username + "无权限"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NoPermission)
		return
	}
	//解析时间转换为中国时间(UTC+8)
	ddlTime, err := time.Parse(time.RFC3339, data.Time)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("时间解析失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	// 检查问卷每个题目的序号没有重复且按照顺序递增
	questionNumMap := make(map[int]bool)
	for i, question := range data.Questions {
		if questionNumMap[question.SerialNum] {
			c.Error(&gin.Error{Err: errors.New("题目序号" + strconv.Itoa(question.SerialNum) + "重复"), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.ServerError)
			return
		}
		if i > 0 && question.SerialNum != data.Questions[i-1].SerialNum+1 {
			c.Error(&gin.Error{Err: errors.New("题目序号不按顺序递增"), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.ServerError)
			return
		}
		questionNumMap[question.SerialNum] = true
		question.SerialNum = i + 1
	}
	//修改问卷
	err = service.UpdateSurvey(data.ID, data.Title, data.Desc, data.Img, data.Questions, ddlTime)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("修改问卷失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	utils.JsonSuccessResponse(c, nil)
}

// 删除问卷
type DeleteSurveyData struct {
	ID int `form:"id" binding:"required"`
}

func DeleteSurvey(c *gin.Context) {
	var data DeleteSurveyData
	err := c.ShouldBindQuery(&data)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取参数失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ParamError)
		return
	}
	//鉴权
	user, err := service.GetUserSession(c)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取用户缓存信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NotLogin)
		return
	}
	// 获取问卷
	survey, err := service.GetSurveyByID(data.ID)
	if err == gorm.ErrRecordNotFound {
		c.Error(&gin.Error{Err: errors.New("问卷不存在"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.SurveyNotExist)
		return
	} else if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取问卷信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	//判断权限
	if (user.AdminType != 2) && (user.AdminType != 1 || survey.UserID != user.ID) && !service.UserInManage(user.ID, survey.ID) {
		c.Error(&gin.Error{Err: errors.New(user.Username + "无权限"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NoPermission)
		return
	}
	//删除问卷
	err = service.DeleteSurvey(data.ID)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("删除问卷失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	utils.JsonSuccessResponse(c, nil)
}

// 获取问卷收集数据
type GetSurveyAnswersData struct {
	ID       int    `form:"id" binding:"required"`
	Text     string `form:"text"`
	Unique   bool   `form:"unique"`
	PageNum  int    `form:"page_num" binding:"required"`
	PageSize int    `form:"page_size" binding:"required"`
}

func GetSurveyAnswers(c *gin.Context) {
	var data GetSurveyAnswersData
	err := c.ShouldBindQuery(&data)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取参数失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ParamError)
		return
	}
	//鉴权
	user, err := service.GetUserSession(c)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取用户缓存信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NotLogin)
		return
	}
	// 获取问卷
	survey, err := service.GetSurveyByID(data.ID)
	if err == gorm.ErrRecordNotFound {
		c.Error(&gin.Error{Err: errors.New("问卷不存在"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.SurveyNotExist)
		return
	} else if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取问卷信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	//判断权限
	if (user.AdminType != 2) && (user.AdminType != 1 || survey.UserID != user.ID) && !service.UserInManage(user.ID, survey.ID) {
		c.Error(&gin.Error{Err: errors.New(user.Username + "无权限"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NoPermission)
		return
	}
	//获取问卷收集数据
	var num *int64
	answers, num, err := service.GetSurveyAnswers(data.ID, data.PageNum, data.PageSize, data.Text, data.Unique)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取问卷收集数据失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	utils.JsonSuccessResponse(c, gin.H{
		"answers_data":   answers,
		"total_page_num": math.Ceil(float64(*num) / float64(data.PageSize)),
	})
}

type GetAllSurveyData struct {
	PageNum  int    `form:"page_num" binding:"required"`
	PageSize int    `form:"page_size" binding:"required"`
	Title    string `form:"title"`
}

func GetAllSurvey(c *gin.Context) {
	var data GetAllSurveyData
	err := c.ShouldBindQuery(&data)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取参数失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ParamError)
		return
	}
	user, err := service.GetUserSession(c)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取用户缓存信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NotLogin)
		return
	}
	// 获取问卷
	var response []interface{}
	var surveys []models.Survey
	var totalPageNum *int64
	if user.AdminType == 2 {
		surveys, totalPageNum, err = service.GetAllSurvey(data.PageNum, data.PageSize, data.Title)
		if err != nil {
			c.Error(&gin.Error{Err: errors.New("获取问卷信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.ServerError)
			return
		}
		surveys = service.SortSurvey(surveys)
		response = service.GetSurveyResponse(surveys)
	} else {
		surveys, err = service.GetAllSurveyByUserID(user.ID)
		if err != nil {
			c.Error(&gin.Error{Err: errors.New("获取问卷信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.ServerError)
			return
		}
		managedSurveys, err := service.GetManageredSurveyByUserID(user.ID)
		if err != nil {
			c.Error(&gin.Error{Err: errors.New("获取问卷信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.ServerError)
			return
		}
		for _, manage := range managedSurveys {
			managedSurvey, err := service.GetSurveyByID(manage.SurveyID)
			if err != nil {
				c.Error(&gin.Error{Err: errors.New("获取问卷信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
				utils.JsonErrorResponse(c, code.ServerError)
				return
			}
			surveys = append(surveys, *managedSurvey)
		}
		surveys = service.SortSurvey(surveys)
		response = service.GetSurveyResponse(surveys)
		response, totalPageNum = service.ProcessResponse(response, data.PageNum, data.PageSize, data.Title)
	}

	utils.JsonSuccessResponse(c, gin.H{
		"survey_list":    response,
		"total_page_num": math.Ceil(float64(*totalPageNum) / float64(data.PageSize)),
	})
}

type GetSurveyData struct {
	ID int `form:"id" binding:"required"`
}

type SurveyData struct {
	ID        int            `json:"id"`
	Time      string         `json:"time"`
	Desc      string         `json:"desc"`
	Img       string         `json:"img"`
	Questions []dao.Question `json:"questions"`
}

// 管理员获取问卷题面
func GetSurvey(c *gin.Context) {
	var data GetSurveyData
	err := c.ShouldBindQuery(&data)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取参数失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ParamError)
		return
	}
	user, err := service.GetUserSession(c)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取用户缓存信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NotLogin)
		return
	}
	// 获取问卷
	survey, err := service.GetSurveyByID(data.ID)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取问卷信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	//判断权限
	if (user.AdminType != 2) && (user.AdminType != 1 || survey.UserID != user.ID) && !service.UserInManage(user.ID, survey.ID) {
		c.Error(&gin.Error{Err: errors.New(user.Username + "无权限"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NoPermission)
		return
	}
	// 获取相应的问题
	questions, err := service.GetQuestionsBySurveyID(survey.ID)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取问题失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	// 构建问卷响应
	questionsResponse := make([]map[string]interface{}, 0)
	for _, question := range questions {
		options, err := service.GetOptionsByQuestionID(question.ID)
		if err != nil {
			c.Error(&gin.Error{Err: errors.New("获取选项失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.ServerError)
			return
		}
		optionsResponse := make([]map[string]interface{}, 0)
		for _, option := range options {
			optionResponse := map[string]interface{}{
				"img":        option.Img,
				"content":    option.Content,
				"serial_num": option.SerialNum,
			}
			optionsResponse = append(optionsResponse, optionResponse)
		}
		questionMap := map[string]interface{}{
			"id":            question.SerialNum,
			"serial_num":    question.SerialNum,
			"subject":       question.Subject,
			"description":   question.Description,
			"required":      question.Required,
			"unique":        question.Unique,
			"other_option":  question.OtherOption,
			"img":           question.Img,
			"question_type": question.QuestionType,
			"reg":           question.Reg,
			"options":       optionsResponse,
		}
		questionsResponse = append(questionsResponse, questionMap)
	}
	response := map[string]interface{}{
		"id":        survey.ID,
		"title":     survey.Title,
		"time":      survey.Deadline,
		"desc":      survey.Desc,
		"img":       survey.Img,
		"questions": questionsResponse,
	}

	utils.JsonSuccessResponse(c, response)
}

type DownloadFileData struct {
	ID int `form:"id" binding:"required"`
}

// 下载
func DownloadFile(c *gin.Context) {
	var data DownloadFileData
	err := c.ShouldBindQuery(&data)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取参数失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ParamError)
		return
	}
	user, err := service.GetUserSession(c)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取用户缓存信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NotLogin)
		return
	}
	// 获取问卷
	survey, err := service.GetSurveyByID(data.ID)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取问卷信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	// 判断权限
	if (user.AdminType != 2) && (user.AdminType != 1 || survey.UserID != user.ID) && !service.UserInManage(user.ID, survey.ID) {
		c.Error(&gin.Error{Err: errors.New(user.Username + "无权限"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NoPermission)
		return
	}
	// 获取数据
	answers, err := service.GetAllSurveyAnswers(data.ID)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取问卷收集数据失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	url, err := service.HandleDownloadFile(answers, survey)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("文件下载失败" + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	utils.JsonSuccessResponse(c, url)
}

// 获取统计问卷选择题数据
type GetSurveyStatisticsData struct {
	ID       int `form:"id" binding:"required"`
	PageNum  int `form:"page_num" binding:"required"`
	PageSize int `form:"page_size" binding:"required"`
}

type GetOptionCount struct {
	SerialNum int    `json:"serial_num"` //选项序号
	Content   string `json:"content"`    //选项内容
	Count     int    `json:"count"`      //选项数量
}

type GetSurveyStatisticsResponse struct {
	SerialNum    int              `json:"serial_num"`    //问题序号
	Question     string           `json:"question"`      //问题内容
	QuestionType int              `json:"question_type"` //问题类型  1:单选 2:多选
	Options      []GetOptionCount `json:"options"`       //选项内容
}

func GetSurveyStatistics(c *gin.Context) {
	var data GetSurveyStatisticsData
	if err := c.ShouldBindQuery(&data); err != nil {
		c.Error(&gin.Error{Err: errors.New("获取参数失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ParamError)
		return
	}

	user, err := service.GetUserSession(c)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取用户缓存信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NotLogin)
		return
	}

	survey, err := service.GetSurveyByID(data.ID)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取问卷信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}

	if (user.AdminType != 2) && (user.AdminType != 1 || survey.UserID != user.ID) && !service.UserInManage(user.ID, survey.ID) {
		c.Error(&gin.Error{Err: errors.New(user.Username + "无权限"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NoPermission)
		return
	}

	answersheets, err := service.GetSurveyAnswersBySurveyID(data.ID)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取问卷收集数据失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}

	questionIDs := make([]int, 0)
	for _, sheet := range answersheets {
		for _, answer := range sheet.Answers {
			if answer.QuestionID != 0 {
				questionIDs = append(questionIDs, answer.QuestionID)
			}
		}
	}

	questions, err := service.GetQuestionsByIDs(questionIDs)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取问题信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}

	questionMap := make(map[int]models.Question)
	for _, question := range questions {
		questionMap[question.ID] = question
	}

	optionCounts := make(map[int]map[int]int)
	for _, sheet := range answersheets {
		for _, answer := range sheet.Answers {
			question := questionMap[answer.QuestionID]
			if question.QuestionType == 1 || question.QuestionType == 2 {
				answerOptions := strings.Split(answer.Content, "┋")
				for _, answerOption := range answerOptions {
					option, err := service.GetOptionByQIDAndAnswer(answer.QuestionID, answerOption)
					if err == gorm.ErrRecordNotFound {
						// 则说明是其他选项，计为其他
						if optionCounts[question.ID] == nil {
							optionCounts[question.ID] = make(map[int]int)
						}
						optionCounts[question.ID][0]++
						continue

					} else if err != nil {
						c.Error(&gin.Error{Err: errors.New("获取选项信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
						utils.JsonErrorResponse(c, code.ServerError)
						return
					}
					if optionCounts[question.ID] == nil {
						optionCounts[question.ID] = make(map[int]int)
					}
					optionCounts[question.ID][option.SerialNum]++
				}
			}
		}
	}
	response := make([]GetSurveyStatisticsResponse, 0, len(optionCounts))
	for qid, options := range optionCounts {
		q := questionMap[qid]
		var qOptions []GetOptionCount
		if q.OtherOption {
			qOptions = make([]GetOptionCount, 0, len(options)+1)
			// 添加其他选项
			qOptions = append(qOptions, GetOptionCount{
				SerialNum: 0,
				Content:   "其他",
				Count:     options[0],
			})
		} else {
			qOptions = make([]GetOptionCount, 0, len(options))
		}

		// 按序号排序
		sortedSerialNums := make([]int, 0, len(options))
		for oSerialNum := range options {
			sortedSerialNums = append(sortedSerialNums, oSerialNum)
		}
		sort.Ints(sortedSerialNums)
		for _, oSerialNum := range sortedSerialNums {
			if oSerialNum == 0 {
				continue
			}
			count := options[oSerialNum]
			op, err := service.GetOptionByQIDAndSerialNum(q.ID, oSerialNum)
			if err != nil {
				c.Error(&gin.Error{Err: errors.New("获取选项信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
				utils.JsonErrorResponse(c, code.ServerError)
				return
			}
			qOptions = append(qOptions, GetOptionCount{
				SerialNum: op.SerialNum,
				Content:   op.Content,
				Count:     count,
			})
		}
		response = append(response, GetSurveyStatisticsResponse{
			SerialNum:    q.SerialNum,
			Question:     q.Subject,
			QuestionType: q.QuestionType,
			Options:      qOptions,
		})
	}
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

	// 按序号排序
	sort.Slice(response, func(i, j int) bool {
		return response[i].SerialNum < response[j].SerialNum
	})

	// 访问切片
	resp := response[start:end]
	totalSumPage := math.Ceil(float64(len(response)) / float64(data.PageSize))

	utils.JsonSuccessResponse(c, gin.H{
		"statistics":     resp,
		"total":          len(answersheets),
		"total_sum_page": totalSumPage,
	})
}

type GetQuestionPreData struct {
	Type string `form:"type"`
}

func GetQuestionPre(c *gin.Context) {
	var data GetQuestionPreData
	if err := c.ShouldBindQuery(&data); err != nil {
		c.Error(&gin.Error{Err: errors.New("获取参数失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ParamError)
		return
	}

	user, err := service.GetUserSession(c)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取用户缓存信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NotLogin)
		return
	}

	if (user.AdminType != 2) && (user.AdminType != 1) {
		c.Error(&gin.Error{Err: errors.New(user.Username + "无权限"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NoPermission)
		return
	}

	// 获取预先信息
	value, err := service.GetQuestionPre(data.Type)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取预先信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	utils.JsonSuccessResponse(c, gin.H{
		"value": value,
	})
}

type CreateQuestionPreData struct {
	Type  string   `json:"type"`
	Value []string `json:"value"`
}

func CreateQuestionPre(c *gin.Context) {
	var data CreateQuestionPreData
	if err := c.ShouldBindJSON(&data); err != nil {
		c.Error(&gin.Error{Err: errors.New("获取参数失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ParamError)
		return
	}

	user, err := service.GetUserSession(c)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取用户缓存信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NotLogin)
		return
	}

	if (user.AdminType != 2) && (user.AdminType != 1) {
		c.Error(&gin.Error{Err: errors.New(user.Username + "无权限"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NoPermission)
		return
	}

	// 创建预先信息
	err = service.CreateQuestionPre(data.Type, data.Value)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("创建预先信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	utils.JsonSuccessResponse(c, nil)
}
