package model

// Manage 问卷权限模型
type Manage struct {
	ID       int    `json:"id"`
	UserID   int    `json:"user_id"`
	SurveyID string `json:"survey_id"`
}
