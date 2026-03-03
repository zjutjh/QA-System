package user

import (
	"errors"
	"sort"
	"strings"

	"QA-System/internal/model"
	"QA-System/internal/pkg/code"
	"QA-System/internal/pkg/utils"
	"QA-System/internal/service"

	"github.com/gin-gonic/gin"
)

type getOptionCount struct {
	SerialNum int    `json:"serial_num"` // 选项序号
	Content   string `json:"content"`    // 选项内容
	Count     int    `json:"count"`      // 选项数量
	Rank      int    `json:"rank"`       // 选项排名
}

type getSurveyStatisticsResponse struct {
	SerialNum    int              `json:"serial_num"`    // 问题序号
	Question     string           `json:"question"`      // 问题内容
	QuestionType int              `json:"question_type"` // 问题类型  1:单选 2:多选
	Options      []getOptionCount `json:"options"`       // 选项内容
}

// GetSurveyStatistics 获取投票统计
func GetSurveyStatistics(c *gin.Context) {
	var data getSurveyData
	err := c.ShouldBindQuery(&data)
	if err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	survey, err := service.GetSurveyByID(data.ID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	if survey.Type != 1 {
		code.AbortWithException(c, code.SurveyTypeError, errors.New("问卷为调研问卷"))
		return
	}
	answerSheets, err := service.GetSurveyAnswersBySurveyID(data.ID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}

	questions, err := service.GetQuestionsBySurveyID(data.ID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}

	// 如果 answerSheets 为空，则返回所有问题和选项统计为 0
	if len(answerSheets) == 0 {
		response := make([]getSurveyStatisticsResponse, 0, len(questions))
		for _, q := range questions {
			options, err := service.GetOptionsByQuestionID(q.ID)
			if err != nil {
				code.AbortWithException(c, code.ServerError, err)
				return
			}

			qOptions := make([]getOptionCount, 0, len(options)+1)
			for _, option := range options {
				qOptions = append(qOptions, getOptionCount{
					SerialNum: option.SerialNum,
					Content:   option.Content,
					Count:     0,
					Rank:      1,
				})
			}

			// 如果支持 "其他" 选项，添加一项
			if q.OtherOption {
				qOptions = append(qOptions, getOptionCount{
					SerialNum: 0,
					Content:   "其他",
					Count:     0,
					Rank:      1,
				})
			}

			response = append(response, getSurveyStatisticsResponse{
				SerialNum:    q.SerialNum,
				Question:     q.Subject,
				QuestionType: q.QuestionType,
				Options:      qOptions,
			})
		}
		utils.JsonSuccessResponse(c, gin.H{"statistics": response})
		return
	}

	// 问题编号对应的问题
	questionMap := make(map[int]model.Question)
	// 问题编号对应的选项们
	optionsMap := make(map[int][]model.Option)
	// 问题编号与选项内容对应的选项
	optionAnswerMap := make(map[int]map[string]model.Option)
	// 问题编号与选项序号对应的选项
	optionSerialNumMap := make(map[int]map[int]model.Option)
	for _, question := range questions {
		questionMap[question.ID] = question
		optionAnswerMap[question.ID] = make(map[string]model.Option)
		optionSerialNumMap[question.ID] = make(map[int]model.Option)
		options, err := service.GetOptionsByQuestionID(question.ID)
		if err != nil {
			code.AbortWithException(c, code.ServerError, err)
			return
		}
		optionsMap[question.ID] = options
		for _, option := range options {
			optionAnswerMap[question.ID][option.Content] = option
			optionSerialNumMap[question.ID][option.SerialNum] = option
		}
	}

	// 问题编号对应的选项编号对应的选项数量
	optionCounts := make(map[int]map[int]int)
	for _, sheet := range answerSheets {
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
			if question.QuestionType == 1 {
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

	response := make([]getSurveyStatisticsResponse, 0, len(optionCounts))
	for qid, options := range optionCounts {
		q := questionMap[qid]
		var qOptions []getOptionCount
		if q.OtherOption {
			qOptions = make([]getOptionCount, 0, len(options)+1)
			// 添加其他选项
			qOptions = append(qOptions, getOptionCount{
				SerialNum: 0,
				Content:   "其他",
				Count:     options[0],
			})
		} else {
			qOptions = make([]getOptionCount, 0, len(options))
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
			qOptions = append(qOptions, getOptionCount{
				SerialNum: op.SerialNum,
				Content:   op.Content,
				Count:     count,
			})
		}

		// 创建一个副本用于排序
		sortedQOptions := make([]getOptionCount, len(qOptions))
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

		response = append(response, getSurveyStatisticsResponse{
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
