package adminController

import (
	"QA-system/apiException"
	"QA-system/controller/req"
	"QA-system/service/listService"
	"QA-system/utility"
	"github.com/gin-gonic/gin"
)

func CreateList(c *gin.Context) {
	var postForm req.CreateListData
	err := c.ShouldBindJSON(&postForm)
	if err != nil {
		_ = c.AbortWithError(200, apiException.ParamError)
		return
	}
	err = listService.CreateList(postForm)
	if err != nil {
		_ = c.AbortWithError(500, apiException.ServerError)
		return
	}
	utility.JsonSuccessResponse(c, nil)
}
