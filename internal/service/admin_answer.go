package service

import (
	"QA-System/internal/dao"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetSurveyAnswers 获取问卷答案（分页）
func GetSurveyAnswers(id int64, num int, size int, text string, unique bool) (dao.AnswersResonse, *int64, error) {
	var answerSheets []dao.AnswerSheet
	data := make([]dao.QuestionAnswers, 0)
	times := make([]string, 0)
	aids := make([]primitive.ObjectID, 0)
	var total *int64
	// 获取问题
	questions, err := d.GetQuestionsBySurveyID(ctx, id)
	if err != nil {
		return dao.AnswersResonse{}, nil, err
	}
	// 初始化data
	questionIndex := make(map[int]int, len(questions))
	for _, question := range questions {
		var q dao.QuestionAnswers
		q.Title = question.Subject
		q.QuestionType = question.QuestionType
		q.Answers = make([]string, 0)
		data = append(data, q)
		questionIndex[question.ID] = len(data) - 1
	}
	// 获取答卷
	answerSheets, total, err = d.GetAnswerSheetBySurveyID(ctx, id, num, size, text, unique)
	if err != nil {
		return dao.AnswersResonse{}, nil, err
	}
	// 填充data
	for _, answerSheet := range answerSheets {
		times = append(times, answerSheet.Time)
		aids = append(aids, answerSheet.AnswerID)
		for _, answer := range answerSheet.Answers {
			if i, ok := questionIndex[answer.QuestionID]; ok {
				data[i].Answers = append(data[i].Answers, answer.Content)
			}
		}
	}
	return dao.AnswersResonse{QuestionAnswers: data, AnswerIDs: aids, Time: times}, total, nil
}

// GetAllSurveyAnswers 获取所有问卷答案
func GetAllSurveyAnswers(id int64) (dao.AnswersResonse, error) {
	data := make([]dao.QuestionAnswers, 0)
	times := make([]string, 0)
	questions, err := d.GetQuestionsBySurveyID(ctx, id)
	if err != nil {
		return dao.AnswersResonse{}, err
	}
	questionIndex := make(map[int]int, len(questions))
	for _, question := range questions {
		var q dao.QuestionAnswers
		q.Title = question.Subject
		q.QuestionType = question.QuestionType
		data = append(data, q)
		questionIndex[question.ID] = len(data) - 1
	}
	answerSheets, _, err := d.GetAnswerSheetBySurveyID(ctx, id, 0, 0, "", true)
	if err != nil {
		return dao.AnswersResonse{}, err
	}
	for _, answerSheet := range answerSheets {
		times = append(times, answerSheet.Time)
		for _, answer := range answerSheet.Answers {
			if i, ok := questionIndex[answer.QuestionID]; ok {
				data[i].Answers = append(data[i].Answers, answer.Content)
			}
		}
	}
	return dao.AnswersResonse{QuestionAnswers: data, Time: times}, nil
}

// GetSurveyAnswersBySurveyID 根据问卷编号获取问卷答案
func GetSurveyAnswersBySurveyID(sid int64) ([]dao.AnswerSheet, error) {
	answerSheets, _, err := d.GetAnswerSheetBySurveyID(ctx, sid, 0, 0, "", true)
	return answerSheets, err
}

// DeleteAnswerSheetBySurveyID 根据问卷编号删除问卷答案
func DeleteAnswerSheetBySurveyID(surveyID int64) error {
	err := d.DeleteAnswerSheetBySurveyID(ctx, surveyID)
	return err
}

// DeleteAnswerSheetByAnswerID 根据答卷ID删除答卷
func DeleteAnswerSheetByAnswerID(answerID primitive.ObjectID) error {
	err := d.DeleteAnswerSheetByAnswerID(ctx, answerID)
	return err
}

// GetAnswerSheetByAnswerID 根据答卷ID获取答卷
func GetAnswerSheetByAnswerID(answerID primitive.ObjectID) error {
	err := d.GetAnswerSheetByAnswerID(ctx, answerID)
	return err
}

// DeleteOauthRecord 删除统一记录
func DeleteOauthRecord(sid int64) error {
	return d.DeleteRecordSheets(ctx, sid)
}
