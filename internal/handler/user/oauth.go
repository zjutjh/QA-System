package user

import (
	"errors"

	"QA-System/internal/pkg/code"
	"QA-System/internal/pkg/utils"
	"QA-System/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/zjutjh/WeJH-SDK/oauth/oauthException"
)

type oauthData struct {
	StudentID string `json:"stu_id" binding:"required"`
	Password  string `json:"password" binding:"required"`
	ID        int64  `json:"id" binding:"required"`
}

// Oauth 统一验证
func Oauth(c *gin.Context) {
	var data oauthData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	user, err := service.Oauth(data.StudentID, data.Password)
	if err != nil {
		var oauthErr *oauthException.Error
		if !errors.As(err, &oauthErr) {
			code.AbortWithException(c, code.ServerError, err)
			return
		}

		switch {
		case errors.Is(oauthErr, oauthException.WrongPassword),
			errors.Is(oauthErr, oauthException.WrongAccount):
			code.AbortWithException(c, code.WrongOauthUsernameOrPassword, err)
		case errors.Is(oauthErr, oauthException.ClosedError):
			code.AbortWithException(c, code.OauthTimeError, err)
		default:
			code.AbortWithException(c, code.ServerError, err)
		}
		return
	}
	token := utils.NewJWT(user.Name, user.College, user.StudentID, user.UserType, user.UserTypeDesc, user.Gender)
	if token == "" {
		code.AbortWithException(c, code.ServerError, errors.New("统一验证失败原因: token生成失败"))
		return
	}
	survey, err := service.GetSurveyByID(data.ID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	dailyLimit, err := service.GetUserLimit(c, data.StudentID, survey.ID, "dailyLimit")
	if err != nil && !errors.Is(err, redis.Nil) {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	sumLimit, err := service.GetUserLimit(c, data.StudentID, survey.ID, "sumLimit")
	if err != nil && !errors.Is(err, redis.Nil) {
		code.AbortWithException(c, code.ServerError, err)
		return
	}

	utils.JsonSuccessResponse(c, gin.H{
		"token":      token,
		"daily_left": survey.DailyLimit - dailyLimit,
		"sum_left":   survey.SumLimit - sumLimit,
	})
}
