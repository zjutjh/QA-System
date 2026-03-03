package service

import (
	"QA-System/internal/model"
)

// CreatePermission 创建权限
func CreatePermission(id int, surveyID int64) error {
	err := d.CreateManage(ctx, id, surveyID)
	return err
}

// DeletePermission 删除权限
func DeletePermission(id int, surveyID int64) error {
	err := d.DeleteManage(ctx, id, surveyID)
	return err
}

// CheckPermission 检查权限
func CheckPermission(id int, surveyID int64) error {
	err := d.CheckManage(ctx, id, surveyID)
	return err
}

// UserInManage 用户是否在管理中
func UserInManage(uid int, sid int64) bool {
	_, err := d.GetManageByUIDAndSID(ctx, uid, sid)
	return err == nil
}

// GetManagedSurveyByUserID 获取用户管理的问卷
func GetManagedSurveyByUserID(userId int) ([]model.Manage, error) {
	var manages []model.Manage
	manages, err := d.GetManageByUserID(ctx, userId)
	return manages, err
}
