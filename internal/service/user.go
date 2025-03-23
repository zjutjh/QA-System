package service

import (
	"bytes"
	"image"
	_ "image/gif" // 注册解码器
	"image/jpeg"
	_ "image/png" // 注册解码器
	"io"
	"os"
	"path/filepath"
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

// GetSurveyByID 根据ID获取问卷
func GetSurveyByID(id int) (*model.Survey, error) {
	survey, err := d.GetSurveyByID(ctx, id)
	return survey, err
}

// GetQuestionsBySurveyID 根据问卷ID获取问题
func GetQuestionsBySurveyID(sid int) ([]model.Question, error) {
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

// GetQuestionByID 根据问卷ID获取问题
func GetQuestionByID(id int) (*model.Question, error) {
	var question *model.Question
	question, err := d.GetQuestionByID(ctx, id)
	return question, err
}

// SubmitSurvey 提交问卷
func SubmitSurvey(sid int, data []dao.QuestionsList, t string) error {
	var answerSheet dao.AnswerSheet
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
	err := d.SaveAnswerSheet(ctx, answerSheet, qids)
	if err != nil {
		return err
	}
	err = d.IncreaseSurveyNum(ctx, sid)
	return err
}

// CreateOauthRecord 创建一条统一验证记录
func CreateOauthRecord(userInfo oauth.UserInfo, t time.Time, sid int) error {
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

// ConvertToJPEG 将图片转换为 JPEG 格式
func ConvertToJPEG(reader io.Reader) (io.Reader, error) {
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 100})
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(buf.Bytes()), nil
}

// SaveFile 保存文件
func SaveFile(reader io.Reader, path string) error {
	dst := filepath.Clean(path)
	err := os.MkdirAll(filepath.Dir(dst), 0750)
	if err != nil {
		return err
	}

	// 创建文件
	outFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func(outFile *os.File) {
		err := outFile.Close()
		if err != nil {
			zap.L().Error("Failed to close file", zap.Error(err))
		}
	}(outFile)

	// 写入文件
	_, err = io.Copy(outFile, reader)
	return err
}

// UpdateVoteLimit 更新投票限制
func UpdateVoteLimit(c *gin.Context, stuId string, surveyID int, isNew bool, durationType string) error {
	if isNew {
		if durationType == "dailyLimit" {
			return SetUserLimit(c, stuId, surveyID, 1, durationType)
		}
		return SetUserSumLimit(c, stuId, surveyID, 1, durationType)
	}
	return InscUserLimit(c, stuId, surveyID, durationType)
}
