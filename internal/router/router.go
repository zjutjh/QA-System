package router

import (
	a "QA-System/internal/handler/admin"
	u "QA-System/internal/handler/user"
	"QA-System/internal/middleware"
	"github.com/gin-gonic/gin"
)

// Init 初始化路由
func Init(r *gin.Engine) {
	const pre = "/api"

	api := r.Group(pre)
	{
		api.POST("/admin/reg", a.Register)
		api.POST("/admin/login", a.Login)
		user := api.Group("/user")
		{
			user.POST("/submit", u.SubmitSurvey)
			user.GET("/get", u.GetSurvey)
			user.GET("/statistic", u.GetSurveyStatistics)
			user.POST("/upload/img", u.UploadImg)
			user.POST("/upload/file", u.UploadFile)
			user.POST("/oauth", u.Oauth)
		}
		admin := api.Group("/admin", middleware.CheckLogin)
		{
			api.POST("/admin/update", a.UpdatePassword)
			api.POST("/admin/reset", a.ResetPassword)
			admin.POST("/create", a.CreateSurvey)
			admin.GET("/create", a.GetQuestionPre)
			admin.POST("/new", a.CreateQuestionPre)
			admin.PUT("/update/status", a.UpdateSurveyStatus)
			admin.PUT("/update/questions", a.UpdateSurvey)
			admin.GET("/list/answers", a.GetSurveyAnswers)
			admin.GET("/statics/answers", a.GetSurveyStatistics)
			admin.DELETE("/delete", a.DeleteSurvey)
			admin.DELETE("/delete/answersheet", a.DeleteAnswerSheet)

			admin.POST("/permission/create", a.CreatePermission)
			admin.DELETE("/permission/delete", a.DeletePermission)

			admin.GET("/list/questions", a.GetAllSurvey)
			admin.GET("/single/question", a.GetSurvey)
			admin.GET("/download", a.DownloadFile)
		}
	}
}
