package user

import (
	"QA-System/internal/dao"
	"QA-System/internal/handler/queue"
	"QA-System/internal/models"
	"QA-System/internal/pkg/code"
	"QA-System/internal/pkg/queue/asynq"
	"QA-System/internal/pkg/utils"
	"QA-System/internal/service"
	"errors"
	"github.com/go-redis/redis/v8"
	"sort"
	"strconv"
	"strings"

	"time"

	"github.com/gin-gonic/gin"
)

type SubmitSurveyData struct {
	ID            int                 `json:"id" binding:"required"`
	Token         string              `json:"token"`
	QuestionsList []dao.QuestionsList `json:"questions_list"`
}

func SubmitSurvey(c *gin.Context) {
	var data SubmitSurveyData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取参数失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ParamError)
		return
	}
	// 判断问卷问题和答卷问题数目是否一致
	survey, err := service.GetSurveyByID(data.ID)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取问卷失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	var stuId string
	if survey.Verify == true {
		stuId, err = utils.ParseJWT(data.Token)
		if err != nil {
			c.Error(&gin.Error{Err: errors.New("获取参数失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.ServerError)
			return
		}
	}
	questions, err := service.GetQuestionsBySurveyID(survey.ID)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取问题失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	if len(questions) != len(data.QuestionsList) {
		c.Error(&gin.Error{Err: errors.New("问卷问题和上传问题数量不一致"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	// 判断填写时间是否在问卷有效期内
	if !survey.Deadline.IsZero() && survey.Deadline.Before(time.Now()) {
		c.Error(&gin.Error{Err: errors.New("填写时间已过"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.TimeBeyondError)
		return
	}
	if !survey.StartTime.IsZero() && survey.StartTime.After(time.Now()) {
		c.Error(&gin.Error{Err: errors.New("填写时间未到"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.TimeBeyondError)
		return
	}
	// 判断问卷是否开放
	if survey.Status != 2 {
		c.Error(&gin.Error{Err: errors.New("问卷未开放"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.SurveyNotOpen)
		return
	}
	// 逐个判断问题答案
	for _, q := range data.QuestionsList {
		question, err := service.GetQuestionByID(q.QuestionID)
		if err != nil {
			c.Error(&gin.Error{Err: errors.New("获取问题失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.ServerError)
			return
		}
		if question.SerialNum != q.SerialNum {
			c.Error(&gin.Error{Err: errors.New("问题序号" + strconv.Itoa(question.ID) + "和" + strconv.Itoa(q.SerialNum) + "不一致"), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.ServerError)
			return
		}
		if question.SurveyID != survey.ID {
			c.Error(&gin.Error{Err: errors.New("问题" + strconv.Itoa(question.SerialNum) + "不属于该问卷"), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.ServerError)
			return
		}
		// 判断必填字段是否为空
		if question.Required && q.Answer == "" {
			c.Error(&gin.Error{Err: errors.New("问题" + strconv.Itoa(q.SerialNum) + "必填字段为空"), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.ServerError)
			return
		}
		// 判断多选题选项数量是否符合要求
		if question.QuestionType == 2 {
			length := len(strings.Split(q.Answer, "┋"))
			if question.MinimumOption != 0 && length < int(question.MinimumOption) {
				c.Error(&gin.Error{Err: errors.New("问题" + strconv.Itoa(q.SerialNum) + "选项数量不符合要求"), Type: gin.ErrorTypeAny})
				utils.JsonErrorResponse(c, code.OptionNumError)
				return
			}
			if question.MaximumOption != 0 && length > int(question.MaximumOption) {
				c.Error(&gin.Error{Err: errors.New("问题" + strconv.Itoa(q.SerialNum) + "选项数量不符合要求"), Type: gin.ErrorTypeAny})
				utils.JsonErrorResponse(c, code.OptionNumError)
				return
			}
		}
	}
	flag := false
	if survey.DailyLimit != 0 && survey.Verify == true {
		limit, err := service.GetUserLimit(c, stuId, survey.ID)
		if err != nil && !errors.Is(err, redis.Nil) {
			c.Error(&gin.Error{Err: errors.New("获取用户投票次数失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.ServerError)
			return
		} else if errors.Is(err, redis.Nil) {
			flag = true
		}
		if err == nil && limit >= int(survey.DailyLimit) {
			c.Error(&gin.Error{Err: errors.New("投票次数已达上限"), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.VoteLimitError)
			return
		}
	}

	if survey.Type != 1 {
		// 创建并入队任务
		task, err := queue.NewSubmitSurveyTask(data.ID, data.QuestionsList)
		if err == redis.Nil {
			c.Error(&gin.Error{Err: errors.New("创建任务失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.StuIDRedisError)
			return
		} else if err != nil {
			c.Error(&gin.Error{Err: errors.New("创建任务失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.ServerError)
			return
		}
		_, err = asynq.Client.Enqueue(task)
		if err != nil {
			c.Error(&gin.Error{Err: errors.New("任务入队失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.ServerError)
			return
		}

	} else {
		err = service.SubmitSurvey(data.ID, data.QuestionsList, time.Now().Format("2006-01-02 15:04:05"))
		if err != nil {
			c.Error(&gin.Error{Err: errors.New("提交问卷失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.ServerError)
			return
		}
		err = service.InscUserLimit(c, stuId, survey.ID)
		if err != nil {
			c.Error(&gin.Error{Err: errors.New("更新用户投票次数失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.ServerError)
			return
		}
	}

	if survey.Verify == true && survey.DailyLimit != 0 {
		if flag {
			err = service.SetUserLimit(c, stuId, survey.ID, 0)
			if err != nil {
				c.Error(&gin.Error{Err: errors.New("设置用户投票次数失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
				utils.JsonErrorResponse(c, code.ServerError)
				return
			}
		}
		err = service.InscUserLimit(c, stuId, survey.ID)
		if err != nil {
			c.Error(&gin.Error{Err: errors.New("更新用户投票次数失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.ServerError)
			return
		}
	} else if survey.Verify == true {
		err = service.CreateOauthRecord(stuId, time.Now(), data.ID)
		if err != nil {
			c.Error(&gin.Error{Err: errors.New("统一验证失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.ServerError)
			return
		}
	}
	utils.JsonSuccessResponse(c, nil)
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

// 用户获取问卷
func GetSurvey(c *gin.Context) {
	var data GetSurveyData
	err := c.ShouldBindQuery(&data)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取参数失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ParamError)
		return
	}
	// 获取问卷
	survey, err := service.GetSurveyByID(data.ID)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取问卷失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	// 判断填写时间是否在问卷有效期内
	if !survey.Deadline.IsZero() && survey.Deadline.Before(time.Now()) {
		c.Error(&gin.Error{Err: errors.New("问卷填写时间已截至"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.TimeBeyondError)
		return
	}
	// 判断问卷是否开放
	if survey.Status != 2 {
		c.Error(&gin.Error{Err: errors.New("问卷未开放"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.SurveyNotOpen)
		return
	}
	if survey.StartTime.IsZero() && survey.StartTime.After(time.Now()) {
		c.Error(&gin.Error{Err: errors.New("问卷未开放"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.SurveyNotOpen)
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
				"img":         option.Img,
				"content":     option.Content,
				"description": option.Description,
				"serial_num":  option.SerialNum,
			}
			optionsResponse = append(optionsResponse, optionResponse)
		}
		questionMap := map[string]interface{}{
			"id":             question.ID,
			"serial_num":     question.SerialNum,
			"subject":        question.Subject,
			"describe":       question.Description,
			"required":       question.Required,
			"unique":         question.Unique,
			"other_option":   question.OtherOption,
			"img":            question.Img,
			"question_type":  question.QuestionType,
			"reg":            question.Reg,
			"maximum_option": question.MaximumOption,
			"minimum_option": question.MinimumOption,
			"options":        optionsResponse,
		}
		questionsResponse = append(questionsResponse, questionMap)
	}
	response := map[string]interface{}{
		"id":          survey.ID,
		"title":       survey.Title,
		"time":        survey.Deadline,
		"desc":        survey.Desc,
		"img":         survey.Img,
		"daily_limit": survey.DailyLimit,
		"verify":      survey.Verify,
		"survey_type": survey.Type,
		"start_time":  survey.StartTime,
		"questions":   questionsResponse,
	}

	utils.JsonSuccessResponse(c, response)
}

// 上传图片
func UploadImg(c *gin.Context) {
	url, err := service.HandleImgUpload(c)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("上传图片失败" + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	utils.JsonSuccessResponse(c, url)
}

// 上传文件
func UploadFile(c *gin.Context) {
	url, err := service.HandleFileUpload(c)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("上传文件失败" + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	utils.JsonSuccessResponse(c, url)
}

type OauthData struct {
	StudentID string `json:"stu_id" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

func Oauth(c *gin.Context) {
	var data OauthData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("统一验证失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ParamError)
		return
	}
	err = service.Oauth(data.StudentID, data.Password)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("统一验证失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		if apiErr, ok := err.(*code.Error); ok {
			utils.JsonErrorResponse(c, apiErr)
		} else {
			utils.JsonErrorResponse(c, code.ServerError)
		}
		return
	}
	token := utils.NewJWT(data.StudentID)
	if token == "" {
		c.Error(&gin.Error{Err: errors.New("统一验证失败原因: token生成失败"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	utils.JsonSuccessResponse(c, gin.H{"token": token})
}

type GetOptionCount struct {
	SerialNum int    `json:"serial_num"` //选项序号
	Content   string `json:"content"`    //选项内容
	Count     int    `json:"count"`      //选项数量
	Rank      int    `json:"rank"`       //选项排名
}

type GetSurveyStatisticsResponse struct {
	SerialNum    int              `json:"serial_num"`    //问题序号
	Question     string           `json:"question"`      //问题内容
	QuestionType int              `json:"question_type"` //问题类型  1:单选 2:多选
	Options      []GetOptionCount `json:"options"`       //选项内容
}

func GetSurveyStatistics(c *gin.Context) {
	var data GetSurveyData
	err := c.ShouldBindQuery(&data)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取参数失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ParamError)
		return
	}
	survey, err := service.GetSurveyByID(data.ID)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取问卷失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	if survey.Type != 1 {
		c.Error(&gin.Error{Err: errors.New("问卷为调研问卷"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.SurveyTypeError)
		return
	}
	answersheets, err := service.GetSurveyAnswersBySurveyID(data.ID)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取问卷收集数据失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}

	questions, err := service.GetQuestionsBySurveyID(data.ID)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取问题信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}

	// 如果 answersheets 为空，则返回所有问题和选项统计为 0
	if len(answersheets) == 0 {
		response := make([]GetSurveyStatisticsResponse, 0, len(questions))
		for _, q := range questions {
			options, err := service.GetOptionsByQuestionID(q.ID)
			if err != nil {
				c.Error(&gin.Error{Err: errors.New("获取选项信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
				utils.JsonErrorResponse(c, code.ServerError)
				return
			}

			qOptions := make([]GetOptionCount, 0, len(options)+1)
			for _, option := range options {
				qOptions = append(qOptions, GetOptionCount{
					SerialNum: option.SerialNum,
					Content:   option.Content,
					Count:     0,
					Rank:      1,
				})
			}

			// 如果支持 "其他" 选项，添加一项
			if q.OtherOption {
				qOptions = append(qOptions, GetOptionCount{
					SerialNum: 0,
					Content:   "其他",
					Count:     0,
					Rank:      1,
				})
			}

			response = append(response, GetSurveyStatisticsResponse{
				SerialNum:    q.SerialNum,
				Question:     q.Subject,
				QuestionType: q.QuestionType,
				Options:      qOptions,
			})
		}
		utils.JsonSuccessResponse(c, gin.H{"statistics": response})
		return
	}

	questionMap := make(map[int]models.Question)
	optionsMap := make(map[int][]models.Option)
	optionAnswerMap := make(map[int]map[string]models.Option)
	optionSerialNumMap := make(map[int]map[int]models.Option)
	for _, question := range questions {
		questionMap[question.ID] = question
		optionAnswerMap[question.ID] = make(map[string]models.Option)
		optionSerialNumMap[question.ID] = make(map[int]models.Option)
		options, err := service.GetOptionsByQuestionID(question.ID)
		if err != nil {
			c.Error(&gin.Error{Err: errors.New("获取选项信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.ServerError)
			return
		}
		optionsMap[question.ID] = options
		for _, option := range options {
			optionAnswerMap[question.ID][option.Content] = option
			optionSerialNumMap[question.ID][option.SerialNum] = option
		}
	}

	optionCounts := make(map[int]map[int]int)
	for _, sheet := range answersheets {
		for _, answer := range sheet.Answers {
			options := optionsMap[answer.QuestionID]
			question := questionMap[answer.QuestionID]
			// 初始化选项统计（确保每个选项的计数存在且为 0）
			if _, initialized := optionCounts[question.ID]; !initialized {
				counts := ensureMap(optionCounts, question.ID)
				for _, option := range options {
					counts[option.SerialNum] = 0
				}
			}
			if question.QuestionType == 1 || question.QuestionType == 2 {
				answerOptions := strings.Split(answer.Content, "┋")
				questionOptions := optionAnswerMap[answer.QuestionID]
				for _, answerOption := range answerOptions {
					// 查找选项
					if questionOptions != nil {
						option, exists := questionOptions[answerOption]
						if exists {
							// 如果找到选项，处理逻辑
							ensureMap(optionCounts, answer.QuestionID)[option.SerialNum]++
							continue
						}
					}
					// 如果选项不存在，处理为 "其他" 选项
					ensureMap(optionCounts, answer.QuestionID)[0]++
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
			count := options[oSerialNum]
			op := optionSerialNumMap[qid][oSerialNum]
			qOptions = append(qOptions, GetOptionCount{
				SerialNum: op.SerialNum,
				Content:   op.Content,
				Count:     count,
			})
		}

		// 创建一个副本用于排序
		sortedQOptions := make([]GetOptionCount, len(qOptions))
		copy(sortedQOptions, qOptions)

		// 按选项数量排序
		sort.Slice(sortedQOptions, func(i, j int) bool {
			// 按数量降序排列，数量相同时按序号升序排列
			if sortedQOptions[i].Count == sortedQOptions[j].Count {
				return sortedQOptions[i].SerialNum < sortedQOptions[j].SerialNum
			}
			return sortedQOptions[i].Count > sortedQOptions[j].Count
		})

		// 补充 rank
		rankMap := make(map[int]int) // 用于记录选项的排名
		currentRank := 1
		for i := 0; i < len(sortedQOptions); i++ {
			if i > 0 && sortedQOptions[i].Count < sortedQOptions[i-1].Count {
				// 当前排名等于前面所有项目数量
				currentRank = i + 1
			}
			rankMap[sortedQOptions[i].SerialNum] = currentRank
		}

		// 将排名写回原始的 qOptions
		for i := range qOptions {
			qOptions[i].Rank = rankMap[qOptions[i].SerialNum]
		}

		response = append(response, GetSurveyStatisticsResponse{
			SerialNum:    q.SerialNum,
			Question:     q.Subject,
			QuestionType: q.QuestionType,
			Options:      qOptions,
		})

	}
	utils.JsonSuccessResponse(c, gin.H{"statistics": response})
}

func ensureMap(m map[int]map[int]int, key int) map[int]int {
	if m[key] == nil {
		m[key] = make(map[int]int)
	}
	return m[key]
}
