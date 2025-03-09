package dao

import (
	"context"

	"QA-System/internal/model"
)

// CreateManage 创建问卷权限
func (d *Dao) CreateManage(ctx context.Context, id int, surveyID string) error {
	err := d.orm.WithContext(ctx).Create(&model.Manage{UserID: id, SurveyID: surveyID}).Error
	return err
}

// DeleteManage 删除问卷权限
func (d *Dao) DeleteManage(ctx context.Context, id int, surveyID string) error {
	err := d.orm.WithContext(ctx).Where("user_id = ? AND survey_id = ?", id, surveyID).Delete(&model.Manage{}).Error
	return err
}

// DeleteManageBySurveyID 根据问卷ID删除问卷权限
func (d *Dao) DeleteManageBySurveyID(ctx context.Context, surveyID string) error {
	err := d.orm.WithContext(ctx).Where("survey_id = ?", surveyID).Delete(&model.Manage{}).Error
	return err
}

// CheckManage 检查问卷权限
func (d *Dao) CheckManage(ctx context.Context, id int, surveyID string) error {
	var manage model.Manage
	err := d.orm.WithContext(ctx).Where("user_id = ? AND survey_id = ?", id, surveyID).First(&manage).Error
	return err
}

// GetManageByUIDAndSID 根据用户ID和问卷UUID获取问卷权限
func (d *Dao) GetManageByUIDAndSID(ctx context.Context, uid int, sid string) (*model.Manage, error) {
	var manage model.Manage
	err := d.orm.WithContext(ctx).Where("user_id = ? AND survey_id = ?", uid, sid).First(&manage).Error
	return &manage, err
}

// GetManageByUserID 根据用户ID获取问卷权限
func (d *Dao) GetManageByUserID(ctx context.Context, uid int) ([]model.Manage, error) {
	var manages []model.Manage
	err := d.orm.WithContext(ctx).Where("user_id = ?", uid).Find(&manages).Error
	return manages, err
}
