package router

import (
	"QA-system/controller/userController"
	"github.com/gin-gonic/gin"
)

func userRouterInit(r *gin.RouterGroup) {
	user := r.Group("/user")
	{
		user.POST("/login", userController.Login)
		user.POST("/add", userController.AddAns)
		user.GET("/get", userController.GetListByID)
		user.GET("/all/get", userController.GetAllList)
	}
}
