package adminController

import (
	"QA-system/apiException"
	"QA-system/service/ansService"
	"QA-system/service/listService"
	"QA-system/utility"
	"github.com/gin-gonic/gin"
	"strconv"
)

func GetAdminListByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		_ = c.AbortWithError(500, apiException.ParamError)
		return
	}
	data, err := listService.GetListByID(id)
	if err != nil {
		_ = c.AbortWithError(500, apiException.ServerError)
		return
	}
	utility.JsonSuccessResponse(c, gin.H{
		"title":   data.List.Title,
		"draft":   data.List.Draft,
		"public":  data.List.Public,
		"number":  data.List.Num,
		"list_id": data.List.ID,
		"list":    data.Question,
	})
}

func GetAns(c *gin.Context) {
	id, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		_ = c.AbortWithError(500, apiException.ParamError)
		return
	}
	list, err := ansService.GetAns(id)
	if err != nil {
		_ = c.AbortWithError(500, apiException.ServerError)
		return
	}
	utility.JsonSuccessResponse(c, list)
}

func GetList(c *gin.Context) {
	list, err := listService.GetAll()
	if err != nil {
		_ = c.AbortWithError(500, apiException.ServerError)
		return
	}
	utility.JsonSuccessResponse(c, list)
}
