package service

import (
	_ "image/gif" // 注册解码器
	_ "image/png" // 注册解码器
	"sort"
	"strings"
	"sync"
	"time"

	"QA-System/internal/dao"
	"QA-System/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/WeJH-SDK/oauth"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	_ "golang.org/x/image/bmp" // 注册解码器
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

var uniqueAnswerSheetLocks sync.Map

func getUniqueAnswerSheetLock(sid int64) *sync.Mutex {
	if v, ok := uniqueAnswerSheetLocks.Load(sid); ok {
		if mu, ok := v.(*sync.Mutex); ok {
			return mu
		}
	}
	mu := &sync.Mutex{}
	actual, _ := uniqueAnswerSheetLocks.LoadOrStore(sid, mu)
	return actual.(*sync.Mutex)
}

func saveAnswerSheetWithUniqueLock(sid int64, answerSheet dao.AnswerSheet, qids []int) error {
	if len(qids) == 0 {
		return d.SaveAnswerSheet(ctx, answerSheet, qids)
	}
	// 唯一题判重依赖“查询+更新+插入”组合操作，进程内按问卷串行化可避免并发窗口。
	mu := getUniqueAnswerSheetLock(sid)
	mu.Lock()
	defer mu.Unlock()
	return d.SaveAnswerSheet(ctx, answerSheet, qids)
}

// GetSurveyByID 根据ID获取问卷
func GetSurveyByID(id int64) (*model.Survey, error) {
	survey, err := d.GetSurveyByID(ctx, id)
	return survey, err
}

// GetQuestionsBySurveyID 根据问卷ID获取问题
func GetQuestionsBySurveyID(sid int64) ([]model.Question, error) {
	var questions []model.Question
	questions, err := d.GetQuestionsBySurveyID(ctx, sid)
	return questions, err
}

// GetOptionsByQuestionID 根据问题ID获取选项
func GetOptionsByQuestionID(questionId int) ([]model.Option, error) {
	var options []model.Option
	options, err := d.GetOptionsByQuestionID(ctx, questionId)
	return options, err
}

// GetQuestionByID 根据问题ID获取问题
func GetQuestionByID(id int) (*model.Question, error) {
	var question *model.Question
	question, err := d.GetQuestionByID(ctx, id)
	return question, err
}

// SubmitSurvey 提交问卷
func SubmitSurvey(sid int64, data []dao.QuestionsList, t string, stuId string) error {
	var answerSheet dao.AnswerSheet
	answerSheet.StudentID = stuId
	answerSheet.SurveyID = sid
	answerSheet.Time = t
	answerSheet.Unique = true
	answerSheet.AnswerID = primitive.NewObjectID()
	qids := make([]int, 0)
	for _, q := range data {
		var answer dao.Answer
		question, err := d.GetQuestionByID(ctx, q.QuestionID)
		if err != nil {
			return err
		}
		if question.QuestionType == 3 && question.Unique {
			qids = append(qids, q.QuestionID)
		}
		answer.QuestionID = q.QuestionID
		answer.Content = q.Answer
		answerSheet.Answers = append(answerSheet.Answers, answer)
	}
	err := saveAnswerSheetWithUniqueLock(sid, answerSheet, qids)
	if err != nil {
		return err
	}
	err = d.IncreaseSurveyNum(ctx, sid)
	if err != nil {
		if rollbackErr := d.DeleteAnswerSheetByAnswerID(ctx, answerSheet.AnswerID); rollbackErr != nil {
			zap.L().Error("问卷计数更新失败后回滚答卷失败",
				zap.Int64("survey_id", sid),
				zap.String("answer_id", answerSheet.AnswerID.Hex()),
				zap.Error(rollbackErr),
			)
		}
		return err
	}
	// 通知失败不影响主提交流程，避免客户端因重试造成重复提交。
	if err := FromSurveyIDToMsg(sid); err != nil {
		zap.L().Warn("问卷提交后发送通知失败", zap.Int64("survey_id", sid), zap.Error(err))
	}
	return nil
}

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

// UpdateVoteLimit 更新投票限制
func UpdateVoteLimit(c *gin.Context, stuId string, surveyID int64, isNew bool, durationType string) error {
	if isNew {
		if durationType == "dailyLimit" {
			return SetUserLimit(c, stuId, surveyID, 1, durationType)
		}
		return SetUserSumLimit(c, stuId, surveyID, 1, durationType)
	}
	return InscUserLimit(c, stuId, surveyID, durationType)
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
