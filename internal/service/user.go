package service

import (
	_ "image/gif" // 注册解码器
	_ "image/png" // 注册解码器
	"sync"

	"QA-System/internal/dao"
	"QA-System/internal/model"

	"github.com/gin-gonic/gin"
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
	// 唯一题判重依赖"查询+更新+插入"组合操作，进程内按问卷串行化可避免并发窗口。
	mu := getUniqueAnswerSheetLock(sid)
	mu.Lock()
	defer mu.Unlock()
	return d.SaveAnswerSheet(ctx, answerSheet, qids)
}

// SubmitSurvey 提交问卷
func SubmitSurvey(sid int64, data []model.QuestionSubmit, t string, stuId string) error {
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
