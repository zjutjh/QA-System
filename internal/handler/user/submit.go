package user

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"QA-System/internal/model"
	"QA-System/internal/pkg/code"
	"QA-System/internal/pkg/utils"
	"QA-System/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/WeJH-SDK/oauth"
	"go.uber.org/zap"
)

type submitSurveyData struct {
	ID            int64                  `json:"id" binding:"required"`
	Token         string                 `json:"token"`
	QuestionsList []model.QuestionSubmit `json:"questions_list"`
}

// SubmitSurvey 提交问卷
func SubmitSurvey(c *gin.Context) {
	var data submitSurveyData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	// 判断问卷问题和答卷问题数目是否一致
	survey, err := service.GetSurveyByID(data.ID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	var userInfo oauth.UserInfo
	var stuId string
	if survey.Verify {
		userInfo, err = utils.ParseJWT(data.Token)
		if err != nil {
			code.AbortWithException(c, code.ServerError, err)
			return
		}
		stuId = userInfo.StudentID
	} else {
		stuId = ""
	}
	questions, err := service.GetQuestionsBySurveyID(survey.ID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	if len(questions) != len(data.QuestionsList) {
		code.AbortWithException(c, code.SurveyError, errors.New("问卷问题和上传问题数量不一致"))
		return
	}
	questionMap := make(map[int]model.Question, len(questions))
	for _, question := range questions {
		questionMap[question.ID] = question
	}
	seenQuestions := make(map[int]struct{}, len(data.QuestionsList))
	// 判断填写时间是否在问卷有效期内
	if !survey.Deadline.IsZero() && survey.Deadline.Before(time.Now()) {
		code.AbortWithException(c, code.TimeBeyondError, errors.New("填写时间已过"))
		return
	}
	if !survey.StartTime.IsZero() && survey.StartTime.After(time.Now()) {
		code.AbortWithException(c, code.TimeBeyondError, errors.New("填写时间未到"))
		return
	}
	// 判断问卷是否开放
	if survey.Status != 2 {
		code.AbortWithException(c, code.SurveyNotOpen, errors.New("问卷未开放"))
		return
	}
	// 逐个判断问题答案
	for _, q := range data.QuestionsList {
		question, ok := questionMap[q.QuestionID]
		if !ok {
			code.AbortWithException(c, code.ServerError,
				errors.New("问题"+strconv.Itoa(q.QuestionID)+"不属于该问卷"))
			return
		}
		if _, exists := seenQuestions[q.QuestionID]; exists {
			code.AbortWithException(c, code.SurveyError,
				errors.New("问卷问题重复提交"))
			return
		}
		seenQuestions[q.QuestionID] = struct{}{}
		// 判断必填字段是否为空
		if question.Required && q.Answer == "" {
			code.AbortWithException(c, code.ServerError,
				errors.New("问题"+strconv.Itoa(q.QuestionID)+"必填字段为空"))
			return
		}
		// 判断多选题选项数量是否符合要求
		if (question.QuestionType == 2 && survey.Type == 0) || (question.QuestionType == 1 && survey.Type == 1) {
			length := uint(len(strings.Split(q.Answer, "┋")))
			if question.MinimumOption != 0 && length < question.MinimumOption {
				code.AbortWithException(c, code.OptionNumError, errors.New("问题"+strconv.Itoa(q.QuestionID)+"选项数量不符合要求"))
				return
			}
			if question.MaximumOption != 0 && length > question.MaximumOption {
				code.AbortWithException(c, code.OptionNumError, errors.New("问题"+strconv.Itoa(q.QuestionID)+"选项数量不符合要求"))
				return
			}
		}
	}
	if len(seenQuestions) != len(questionMap) {
		code.AbortWithException(c, code.SurveyError, errors.New("问卷问题缺失或重复"))
		return
	}
	flagSum, flagDay := false, false

	if survey.Verify {
		var err error
		if userInfo.UserTypeDesc != "本科生" && survey.UndergradOnly {
			code.AbortWithException(c, code.NotUnderGraduateError, errors.New("当前问卷仅允许本科生回答"))
			return
		}
		// 统一检查总投票次数和每日投票次数
		if flagSum, err = service.CheckLimit(c, stuId, survey, "sumLimit", survey.SumLimit); err != nil {
			if err.Error() == "sumLimit已达上限" {
				code.AbortWithException(c, code.VoteSumLimitError, errors.New("总投票次数已达上限"))
			} else {
				code.AbortWithException(c, code.ServerError, err)
			}
			return
		}

		if flagDay, err = service.CheckLimit(c, stuId, survey, "dailyLimit", survey.DailyLimit); err != nil {
			if err.Error() == "dailyLimit已达上限" {
				code.AbortWithException(c, code.VoteLimitError, errors.New("单日投票次数已达上限"))
			} else {
				code.AbortWithException(c, code.ServerError, err)
			}
			return
		}
	}

	submitTime := time.Now().Format(time.DateTime)
	err = service.SubmitSurvey(data.ID, data.QuestionsList, submitTime, stuId)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}

	if survey.Verify {
		if survey.DailyLimit > 0 {
			err := service.UpdateVoteLimit(c, stuId, survey.ID, flagDay, "dailyLimit")
			if err != nil {
				zap.L().Warn("问卷提交后更新单日投票限制失败",
					zap.Int64("survey_id", survey.ID),
					zap.String("student_id", stuId),
					zap.Error(err),
				)
			}
		}
		if survey.SumLimit > 0 {
			err := service.UpdateVoteLimit(c, stuId, survey.ID, flagSum, "sumLimit")
			if err != nil {
				zap.L().Warn("问卷提交后更新总投票限制失败",
					zap.Int64("survey_id", survey.ID),
					zap.String("student_id", stuId),
					zap.Error(err),
				)
			}
		}
		// 记录授权
		if err = service.CreateOauthRecord(userInfo, time.Now(), data.ID); err != nil {
			zap.L().Warn("问卷提交后记录统一验证信息失败",
				zap.Int64("survey_id", survey.ID),
				zap.String("student_id", stuId),
				zap.Error(err),
			)
		}
	}
	utils.JsonSuccessResponse(c, gin.H{
		"time": submitTime,
	})
}
