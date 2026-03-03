package service

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"QA-System/internal/dao"
	"QA-System/internal/model"
)

// GetOptionCount 选项数据
type GetOptionCount struct {
	SerialNum int    `json:"serial_num"` // 选项序号
	Content   string `json:"content"`    // 选项内容
	Count     int    `json:"count"`      // 选项数量
	Percent   string `json:"percent"`    // 占比百分比，保留两位小数
}

// GetChooseStatisticsResponse 问题模型
type GetChooseStatisticsResponse struct {
	SerialNum    int              `json:"serial_num"`    // 问题序号
	Question     string           `json:"question"`      // 问题内容
	QuestionType int              `json:"question_type"` // 问题类型  1:单选 2:多选
	Options      []GetOptionCount `json:"options"`       // 选项内容
}

// GenerateQuestionStats 生成问卷题目统计结果
func GenerateQuestionStats(questions []model.Question, answerSheets []dao.AnswerSheet) []GetChooseStatisticsResponse {
	questionMap := make(map[int]model.Question)
	optionsMap := make(map[int][]model.Option)
	optionAnswerMap := make(map[int]map[string]model.Option)
	optionSerialNumMap := make(map[int]map[int]model.Option)
	for _, question := range questions {
		questionMap[question.ID] = question
		optionAnswerMap[question.ID] = make(map[string]model.Option)
		optionSerialNumMap[question.ID] = make(map[int]model.Option)
		options, err := GetOptionsByQuestionID(question.ID)
		if err != nil {
			log.Println("Error fetching options for questionID:", question.ID)
			continue
		}
		optionsMap[question.ID] = options
		for _, option := range options {
			optionAnswerMap[question.ID][option.Content] = option
			optionSerialNumMap[question.ID][option.SerialNum] = option
		}
	}

	optionCounts := make(map[int]map[int]int)
	for _, sheet := range answerSheets {
		for _, answer := range sheet.Answers {
			options := optionsMap[answer.QuestionID]
			question := questionMap[answer.QuestionID]

			// 初始化外层 map：如果某题还没记录，先创建一个 map[int]int 作为它的值
			if _, ok := optionCounts[question.ID]; !ok {
				optionCounts[question.ID] = make(map[int]int)

				// 初始化所有选项的计数为 0（避免后续统计遗漏）
				for _, option := range options {
					optionCounts[question.ID][option.SerialNum] = 0
				}
			}

			// 支持单选题、多选题：用 "┋" 分割用户选的答案
			if question.QuestionType == 1 || question.QuestionType == 2 {
				answerOptions := strings.Split(answer.Content, "┋")
				questionOptions := optionAnswerMap[answer.QuestionID]

				for _, answerOption := range answerOptions {
					if questionOptions != nil {
						option, exists := questionOptions[answerOption]
						if exists {
							// 该选项存在，序号对应的次数 +1
							optionCounts[answer.QuestionID][option.SerialNum]++
							continue
						}
					}
					// "其他"选项，统一用 SerialNum = 0 表示
					optionCounts[answer.QuestionID][0]++
				}
			}
		}
	}
	response := make([]GetChooseStatisticsResponse, 0, len(optionCounts))
	for qid, optionCountMap := range optionCounts {
		q := questionMap[qid]
		if q.QuestionType != 1 && q.QuestionType != 2 {
			continue
		}
		total := 0
		for _, count := range optionCountMap {
			total += count
		}

		var qOptions []GetOptionCount
		if q.OtherOption {
			qOptions = append(qOptions, GetOptionCount{
				SerialNum: 0,
				Content:   "其他",
				Count:     optionCountMap[0],
				Percent:   fmt.Sprintf("%.2f%%", float64(optionCountMap[0])*100/float64(total)),
			})
		}

		// 排序
		serialNums := make([]int, 0, len(optionCountMap))
		for serial := range optionCountMap {
			if serial != 0 {
				serialNums = append(serialNums, serial)
			}
		}
		sort.Ints(serialNums)

		for _, serial := range serialNums {
			count := optionCountMap[serial]
			op := optionSerialNumMap[qid][serial]
			percent := "0.00%"
			if total > 0 {
				percent = fmt.Sprintf("%.2f%%", float64(count)*100/float64(total))
			}
			qOptions = append(qOptions, GetOptionCount{
				SerialNum: serial,
				Content:   op.Content,
				Count:     count,
				Percent:   percent,
			})
		}

		response = append(response, GetChooseStatisticsResponse{
			SerialNum:    q.SerialNum,
			Question:     q.Subject,
			QuestionType: q.QuestionType,
			Options:      qOptions,
		})
	}
	// 按序号排序
	sort.Slice(response, func(i, j int) bool {
		return response[i].SerialNum < response[j].SerialNum
	})
	return response
}
