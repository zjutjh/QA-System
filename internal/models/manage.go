package models

// Manage 问卷权限模型
type Manage struct {
	ID       int `json:"id"`
	UserID   int `json:"user_id"`
	SurveyID int `json:"survey_id"`
}
