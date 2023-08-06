package userController

import (
	"QA-system/apiException"
	"QA-system/service/listService"
	"QA-system/utility"
	"github.com/gin-gonic/gin"
	"strconv"
)

func GetListByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Query("id"))
	data, err := listService.GetListByID(id)
	if err != nil {
		_ = c.AbortWithError(500, apiException.ServerError)
		return
	}
	utility.JsonSuccessResponse(c, gin.H{
		"title": data.List.Title,
		"list":  data.Question,
	})
}

func GetAllList(c *gin.Context) {
	list, err := listService.GetAllPublic()
	if err != nil {
		_ = c.AbortWithError(500, apiException.ServerError)
		return
	}
	utility.JsonSuccessResponse(c, list)
}
