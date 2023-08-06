package router

import (
	"QA-system/controller/adminController"
	midware "QA-system/midware"
	"github.com/gin-gonic/gin"
)

func adminRouterInit(r *gin.RouterGroup) {
	admin := r.Group("/admin", midware.CheckLogin)
	{
		admin.POST("/add", adminController.CreateList)
		admin.GET("/single/get", adminController.GetAdminListByID)
		admin.GET("/detail/get", adminController.GetAns)
		admin.POST("/draft/status", adminController.UpdateStatus)
		admin.GET("/list/get", adminController.GetList)
		admin.POST("/draft/update", adminController.UpdateList)
		admin.POST("/delete", adminController.Delete)
		admin.GET("/download/:list_id", adminController.ExportExcel)
	}
}
