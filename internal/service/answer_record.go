package service

import (
	"sort"
	"strings"
	"time"

	"QA-System/internal/dao"
	"QA-System/internal/model"

	"github.com/zjutjh/WeJH-SDK/oauth"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateOauthRecord 创建一条统一验证记录
func CreateOauthRecord(userInfo oauth.UserInfo, t time.Time, sid int64) error {
	sheet := dao.RecordSheet{
		College:      userInfo.College,
		Name:         userInfo.Name,
		StudentID:    userInfo.StudentID,
		UserType:     userInfo.UserType,
		UserTypeDesc: userInfo.UserTypeDesc,
		Gender:       userInfo.Gender,
		Time:         t,
	}
	return d.SaveRecordSheet(ctx, sheet, sid)
}

// GetSurveyAnswersBySurveyIDAndStudentID 根据问卷编号和学生学号获取问卷答案
func GetSurveyAnswersBySurveyIDAndStudentID(surveyid int64, studentid string) ([]dao.AnswerSheet, error) {
	answerSheets, _, err := d.GetAnswerSheetBySurveyIDAndStudentID(ctx, surveyid, studentid, 0, 0, "", true)
	return answerSheets, err
}

// AnswerRecord 一次提交记录
type AnswerRecord struct {
	AnswerID        primitive.ObjectID    `json:"answer_id"`
	Time            string                `json:"time"`
	SurveyID        int64                 `json:"survey_id"`
	StudentID       string                `json:"student_id"`
	Unique          bool                  `json:"unique"`
	QuestionAnswers []dao.QuestionAnswers `json:"question_answers"`
}

func CreateRecordDetailResponseWithQuestions(userAnswerSheets []dao.AnswerSheet, questions []model.Question) []AnswerRecord {
	qtMap := make(map[int]int, len(questions))
	subMap := make(map[int]string, len(questions))
	for _, q := range questions {
		qid := int(q.ID)
		qtMap[qid] = q.QuestionType
		subMap[qid] = q.Subject
	}

	if len(userAnswerSheets) == 0 {
		return make([]AnswerRecord, 0)
	}
	records := make([]AnswerRecord, 0, len(userAnswerSheets))

	for _, sheet := range userAnswerSheets {
		rec := AnswerRecord{
			AnswerID:        sheet.AnswerID,
			Time:            sheet.Time,
			SurveyID:        sheet.SurveyID,
			StudentID:       sheet.StudentID,
			Unique:          sheet.Unique,
			QuestionAnswers: make([]dao.QuestionAnswers, 0, len(sheet.Answers)),
		}

		// 让每次提交里的题目按题号排一下
		answers := append([]dao.Answer(nil), sheet.Answers...)
		sort.Slice(answers, func(i, j int) bool { return answers[i].SerialNum < answers[j].SerialNum })

		for _, ans := range answers {
			// 清理空串
			parts := strings.Split(ans.Content, "┋")
			clean := make([]string, 0, len(parts))
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					clean = append(clean, p)
				}
			}

			title := subMap[ans.QuestionID]

			rec.QuestionAnswers = append(rec.QuestionAnswers, dao.QuestionAnswers{
				Title:        title,
				QuestionType: qtMap[ans.QuestionID],
				Answers:      clean,
			})
		}

		records = append(records, rec)
	}

	return records
}
