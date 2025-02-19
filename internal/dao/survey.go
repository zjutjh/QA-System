package dao

import (
	"context"
	"time"

	"QA-System/internal/model"
	"gorm.io/gorm"
)

// CreateSurvey 创建问卷
func (d *Dao) CreateSurvey(ctx context.Context, survey model.Survey) (model.Survey, error) {
	err := d.orm.WithContext(ctx).Create(&survey).Error
	return survey, err
}

// UpdateSurveyStatus 更新问卷状态
func (d *Dao) UpdateSurveyStatus(ctx context.Context, surveyID int, status int) error {
	err := d.orm.WithContext(ctx).Model(&model.Survey{}).Where("id = ?", surveyID).Update("status", status).Error
	return err
}

// UpdateSurvey 更新问卷
func (d *Dao) UpdateSurvey(ctx context.Context, id int, surveyType, limit uint,
	verify bool, desc string, title string, deadline, startTime time.Time) error {
	err := d.orm.WithContext(ctx).Model(&model.Survey{}).Where("id = ?", id).
		Updates(model.Survey{
			Deadline:   deadline,
			DailyLimit: limit,
			Verify:     verify,
			Desc:       desc,
			Title:      title,
			Type:       surveyType,
			StartTime:  startTime,
		}).Error
	return err
}

// GetAllSurveyByUserID 获取用户的所有问卷
func (d *Dao) GetAllSurveyByUserID(ctx context.Context, userId int) ([]model.Survey, error) {
	var surveys []model.Survey
	err := d.orm.WithContext(ctx).Model(model.Survey{}).Where("user_id = ?", userId).
		Order("CASE WHEN status = 2 THEN 0 ELSE 1 END, id DESC").Find(&surveys).Error
	return surveys, err
}

// GetSurveyByID 根据问卷ID获取问卷
func (d *Dao) GetSurveyByID(ctx context.Context, surveyID int) (*model.Survey, error) {
	var survey model.Survey
	err := d.orm.WithContext(ctx).Where("id = ?", surveyID).First(&survey).Error
	return &survey, err
}

// GetSurveyByTitle 根据问卷标题获取问卷
func (d *Dao) GetSurveyByTitle(ctx context.Context, title string, num, size int) ([]model.Survey, *int64, error) {
	var surveys []model.Survey
	var sum int64
	err := d.orm.WithContext(ctx).Model(model.Survey{}).
		Where("title like ?", "%"+title+"%").
		Order("CASE WHEN status = 2 THEN 0 ELSE 1 END").
		Count(&sum).Limit(size).Offset((num - 1) * size).Find(&surveys).Error
	return surveys, &sum, err
}

// IncreaseSurveyNum 增加问卷填写人数
func (d *Dao) IncreaseSurveyNum(ctx context.Context, sid int) error {
	err := d.orm.WithContext(ctx).Model(&model.Survey{}).Where("id = ?", sid).
		Update("num", gorm.Expr("num + ?", 1)).Error
	return err
}

// DeleteSurvey 删除问卷
func (d *Dao) DeleteSurvey(ctx context.Context, surveyID int) error {
	err := d.orm.WithContext(ctx).Where("id = ?", surveyID).Delete(&model.Survey{}).Error
	return err
}
