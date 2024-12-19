package dao

import (
	"QA-System/internal/models"
	"context"
	"time"

	"gorm.io/gorm"
)

func (d *Dao) CreateSurvey(ctx context.Context, survey models.Survey) (models.Survey, error) {
	err := d.orm.WithContext(ctx).Create(&survey).Error
	return survey, err
}

func (d *Dao) UpdateSurveyStatus(ctx context.Context, surveyID int, status int) error {
	err := d.orm.WithContext(ctx).Model(&models.Survey{}).Where("id = ?", surveyID).Update("status", status).Error
	return err
}

func (d *Dao) UpdateSurvey(ctx context.Context, id int, surveyType, limit uint, verify bool, title, desc, img string, deadline, startTime time.Time) error {
	err := d.orm.WithContext(ctx).Model(&models.Survey{}).Where("id = ?", id).Updates(models.Survey{Title: title, Desc: desc, Img: img, Deadline: deadline, DailyLimit: limit, Verify: verify, Type: surveyType, StartTime: startTime}).Error
	return err
}

func (d *Dao) GetAllSurveyByUserID(ctx context.Context, userId int) ([]models.Survey, error) {
	var surveys []models.Survey
	err := d.orm.WithContext(ctx).Model(models.Survey{}).Where("user_id = ?", userId).
		Order("CASE WHEN status = 2 THEN 0 ELSE 1 END, id DESC").Find(&surveys).Error
	return surveys, err
}

func (d *Dao) GetSurveyByID(ctx context.Context, surveyID int) (*models.Survey, error) {
	var survey models.Survey
	err := d.orm.WithContext(ctx).Where("id = ?", surveyID).First(&survey).Error
	return &survey, err
}

func (d *Dao) GetSurveyByTitle(ctx context.Context, title string, num, size int) ([]models.Survey, *int64, error) {
	var surveys []models.Survey
	var sum int64
	err := d.orm.WithContext(ctx).Model(models.Survey{}).Where("title like ?", "%"+title+"%").Order("CASE WHEN status = 2 THEN 0 ELSE 1 END").Count(&sum).Limit(size).Offset((num - 1) * size).Find(&surveys).Error
	return surveys, &sum, err
}

func (d *Dao) IncreaseSurveyNum(ctx context.Context, sid int) error {
	err := d.orm.WithContext(ctx).Model(&models.Survey{}).Where("id = ?", sid).Update("num", gorm.Expr("num + ?", 1)).Error
	return err
}

func (d *Dao) DeleteSurvey(ctx context.Context, surveyID int) error {
	err := d.orm.WithContext(ctx).Where("id = ?", surveyID).Delete(&models.Survey{}).Error
	return err
}
