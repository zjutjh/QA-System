package dao

import (
	"QA-System/internal/models"
	"context"
)

type Question struct {
	ID           int      `json:"id"`
	SerialNum    int      `json:"serial_num"`    // 题目序号
	Subject      string   `json:"subject"`       // 问题
	Description  string   `json:"description"`   // 问题描述
	Img          string   `json:"img"`           // 图片
	Required     bool     `json:"required"`      // 是否必填
	Unique       bool     `json:"unique"`        // 是否唯一
	OtherOption  bool     `json:"other_option"`  // 是否有其他选项
	QuestionType int      `json:"question_type" binding:"required,oneof=1 2 3 4 5 6"` // 问题类型 1单选2多选3填空4简答5图片6文件
	Reg          string   `json:"reg"`           // 正则表达式
	Options      []Option `json:"options"`       // 选项
}

type QuestionsList struct {
	QuestionID int    `json:"question_id" binding:"required"`
	SerialNum  int    `json:"serial_num"`
	Answer     string `json:"answer"`
}

func (d *Dao) CreateQuestion(ctx context.Context, question models.Question)(models.Question, error) {
	err := d.orm.WithContext(ctx).Create(&question).Error
	return question,err
}

func (d *Dao) GetQuestionsBySurveyID(ctx context.Context, surveyID int) ([]models.Question, error) {
	var questions []models.Question
	err := d.orm.WithContext(ctx).Model(models.Question{}).Where("survey_id = ?", surveyID).Find(&questions).Error
	return questions, err
}

func (d *Dao) GetQuestionByID(ctx context.Context, questionID int) (*models.Question, error) {
	var question models.Question
	err := d.orm.WithContext(ctx).Where("id = ?", questionID).First(&question).Error
	return &question, err
}

func (d *Dao) DeleteQuestion(ctx context.Context, questionID int) error {
	err := d.orm.WithContext(ctx).Where("id = ?", questionID).Delete(&models.Question{}).Error
	return err
}

func (d *Dao) DeleteQuestionBySurveyID(ctx context.Context, surveyID int) error {
	err := d.orm.WithContext(ctx).Where("survey_id = ?", surveyID).Delete(&models.Question{}).Error
	return err
}

func (d *Dao) CreateType(ctx context.Context,name string,value string)error{
	//如果type已经存在则直接更新当前type的value
	var t models.Type
	err := d.orm.WithContext(ctx).Where("type = ?",name).First(&t).Error
	if err == nil{
		err = d.orm.WithContext(ctx).Model(&t).Update("value",value).Error
		return err
	}
	err = d.orm.WithContext(ctx).Create(&models.Type{Type:name,Value:value}).Error
	return err
}

func (d *Dao) GetType(ctx context.Context,name string)(string,error){
	var t models.Type
	err := d.orm.WithContext(ctx).Where("type = ?",name).First(&t).Error
	return t.Value,err
}