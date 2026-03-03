package admin

import (
	"errors"
	"math"

	"QA-System/internal/pkg/code"
	"QA-System/internal/pkg/utils"
	"QA-System/internal/service"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type getSurveyAnswersData struct {
	ID       int64  `form:"id" binding:"required"`
	Text     string `form:"text"`
	Unique   bool   `form:"unique"`
	PageNum  int    `form:"page_num" binding:"required,gt=0"`
	PageSize int    `form:"page_size" binding:"required,gt=0"`
}

// GetSurveyAnswers 获取问卷收集数据
func GetSurveyAnswers(c *gin.Context) {
	var data getSurveyAnswersData
	err := c.ShouldBindQuery(&data)
	if err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	// 鉴权
	user, err := service.GetUserSession(c)
	if err != nil {
		code.AbortWithException(c, code.NotLogin, err)
		return
	}
	// 获取问卷
	survey, err := service.GetSurveyByID(data.ID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		code.AbortWithException(c, code.SurveyNotExist, errors.New("问卷不存在"))
		return
	} else if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	// 判断权限
	if (user.AdminType != 2) && (user.AdminType != 1 || survey.UserID != user.ID) &&
		!service.UserInManage(user.ID, survey.ID) {
		code.AbortWithException(c, code.NoPermission, errors.New(user.Username+"无权限"))
		return
	}
	// 获取问卷收集数据
	var num *int64
	answers, num, err := service.GetSurveyAnswers(data.ID, data.PageNum, data.PageSize, data.Text, data.Unique)
	if err != nil {
		if err.Error() == "页数超出范围" {
			code.AbortWithException(c, code.PageBeyondError, err)
		} else {
			code.AbortWithException(c, code.ServerError, err)
		}
		return
	}

	utils.JsonSuccessResponse(c, gin.H{
		"answers_data":   answers,
		"survey_type":    survey.Type,
		"total_page_num": math.Ceil(float64(*num) / float64(data.PageSize)),
	})
}

type deleteAnswerSheetData struct {
	AnswerID string `bson:"_id" form:"answer_id" binding:"required"`
}

// DeleteAnswerSheet 删除答卷
func DeleteAnswerSheet(c *gin.Context) {
	var data deleteAnswerSheetData
	err := c.ShouldBindQuery(&data)
	if err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	// 鉴权
	user, err := service.GetUserSession(c)
	if err != nil {
		code.AbortWithException(c, code.NotLogin, err)
		return
	}

	// 将 AnswerID 转换为 ObjectID
	objectID, err := primitive.ObjectIDFromHex(data.AnswerID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	// 获取问卷
	err = service.GetAnswerSheetByAnswerID(objectID)
	if errors.Is(err, mongo.ErrNoDocuments) {
		code.AbortWithException(c, code.AnswerSheetNotExist, errors.New("答卷不存在"))
		return
	} else if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	// 判断权限
	if user.AdminType != 2 {
		code.AbortWithException(c, code.NoPermission, errors.New(user.Username+"无权限"))
		return
	}
	// 删除答卷
	err = service.DeleteAnswerSheetByAnswerID(objectID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	utils.JsonSuccessResponse(c, nil)
}
