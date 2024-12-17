package dao

import (
	"context"

	"QA-System/internal/models"
)

// CreateManage 创建问卷权限
func (d *Dao) CreateManage(ctx context.Context, id int, surveyID int) error {
	err := d.orm.WithContext(ctx).Create(&models.Manage{UserID: id, SurveyID: surveyID}).Error
	return err
}

// DeleteManage 删除问卷权限
func (d *Dao) DeleteManage(ctx context.Context, id int, surveyID int) error {
	err := d.orm.WithContext(ctx).Where("user_id = ? AND survey_id = ?", id, surveyID).Delete(&models.Manage{}).Error
	return err
}

// DeleteManageBySurveyID 根据问卷ID删除问卷权限
func (d *Dao) DeleteManageBySurveyID(ctx context.Context, surveyID int) error {
	err := d.orm.WithContext(ctx).Where("survey_id = ?", surveyID).Delete(&models.Manage{}).Error
	return err
}

// CheckManage 检查问卷权限
func (d *Dao) CheckManage(ctx context.Context, id int, surveyID int) error {
	var manage models.Manage
	err := d.orm.WithContext(ctx).Where("user_id = ? AND survey_id = ?", id, surveyID).First(&manage).Error
	return err
}

// GetManageByUIDAndSID 根据用户ID和问卷ID获取问卷权限
func (d *Dao) GetManageByUIDAndSID(ctx context.Context, uid int, sid int) (*models.Manage, error) {
	var manage models.Manage
	err := d.orm.WithContext(ctx).Where("user_id = ? AND survey_id = ?", uid, sid).First(&manage).Error
	return &manage, err
}

// GetManageByUserID 根据用户ID获取问卷权限
func (d *Dao) GetManageByUserID(ctx context.Context, uid int) ([]models.Manage, error) {
	var manages []models.Manage
	err := d.orm.WithContext(ctx).Where("user_id = ?", uid).Find(&manages).Error
	return manages, err
}
