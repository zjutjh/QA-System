package service

import (
	"time"

	"QA-System/pkg/extension"
)

// FromSurveyIDToMsg 通过问卷ID将问卷信息发送到邮件提醒插件
func FromSurveyIDToMsg(surveyID int64) error {
	// 获取问卷信息
	survey, err := GetSurveyByID(surveyID)
	if err != nil {
		return err
	}

	if survey.NeedNotify {
		creatorEmail, err := GetUserEmailByID(survey.UserID)
		if err != nil {
			return err
		}
		// 构造消息数据
		data := map[string]any{
			"creator_email": creatorEmail,
			"survey_title":  survey.Title,
			"timestamp":     time.Now().UnixNano(),
		}

		// 使用 emailNotifier 发送邮件（安全执行，不会因为插件问题影响主流程）
		extension.ExecutePluginSafely("emailNotifier", data)
	}

	return nil
}
