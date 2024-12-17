package dao

import (
	"context"

	"QA-System/internal/models"
)

func (d *Dao) CreateManage(ctx context.Context, id int, surveyID int) error {
	err := d.orm.WithContext(ctx).Create(&models.Manage{UserID: id, SurveyID: surveyID}).Error
	return err
}

func (d *Dao) DeleteManage(ctx context.Context, id int, surveyID int) error {
	err := d.orm.WithContext(ctx).Where("user_id = ? AND survey_id = ?", id, surveyID).Delete(&models.Manage{}).Error
	return err
}

func (d *Dao) DeleteManageBySurveyID(ctx context.Context, surveyID int) error {
	err := d.orm.WithContext(ctx).Where("survey_id = ?", surveyID).Delete(&models.Manage{}).Error
	return err
}

func (d *Dao) CheckManage(ctx context.Context, id int, surveyID int) error {
	var manage models.Manage
	err := d.orm.WithContext(ctx).Where("user_id = ? AND survey_id = ?", id, surveyID).First(&manage).Error
	return err
}

func (d *Dao) GetManageByUIDAndSID(ctx context.Context, uid int, sid int) (*models.Manage, error) {
	var manage models.Manage
	err := d.orm.WithContext(ctx).Where("user_id = ? AND survey_id = ?", uid, sid).First(&manage).Error
	return &manage, err
}

func (d *Dao) GetManageByUserID(ctx context.Context, uid int) ([]models.Manage, error) {
	var manages []models.Manage
	err := d.orm.WithContext(ctx).Where("user_id = ?", uid).Find(&manages).Error
	return manages, err
}
