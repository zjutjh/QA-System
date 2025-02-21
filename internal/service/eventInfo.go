package service

// 专门负责处理发送到stream的信息的一些函数
import (
	"QA-System/plugins"
	"time"
)

// FromSurveyIDToMsg 通过问卷ID将问卷信息发送到消息队列（直接发给插件好啦）
func FromSurveyIDToMsg(surveyID int) error {
	// 获取问卷信息
	survey, err := GetSurveyByID(surveyID)
	if err != nil {
		return err
	}

	creatorEmail, err1 := GetUserEmailByID(survey.UserID)
	if err1 != nil {
		return err1
	}
	// 构造消息数据
	data := map[string]any{
		"creator_email": creatorEmail,
		"survey_title":  survey.Title,
		"timestamp":     time.Now().UnixNano(),
	}

	// 使用 BetterEmailNotifier 发送邮件
	err = plugins.BetterEmailNotify(data)

	return nil
}

// // 发送到Redis Stream
// err = pkg.PublishToStream(context.Background(), data)
// if err != nil {
// 	return err
// }
