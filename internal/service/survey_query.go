package service

import (
	"QA-System/internal/model"
)

// GetSurveyByID 根据ID获取问卷
func GetSurveyByID(id int64) (*model.Survey, error) {
	survey, err := d.GetSurveyByID(ctx, id)
	return survey, err
}

// GetQuestionsBySurveyID 根据问卷ID获取问题
func GetQuestionsBySurveyID(sid int64) ([]model.Question, error) {
	var questions []model.Question
	questions, err := d.GetQuestionsBySurveyID(ctx, sid)
	return questions, err
}

// GetOptionsByQuestionID 根据问题ID获取选项
func GetOptionsByQuestionID(questionId int) ([]model.Option, error) {
	var options []model.Option
	options, err := d.GetOptionsByQuestionID(ctx, questionId)
	return options, err
}

// GetQuestionByID 根据问题ID获取问题
func GetQuestionByID(id int) (*model.Question, error) {
	var question *model.Question
	question, err := d.GetQuestionByID(ctx, id)
	return question, err
}
