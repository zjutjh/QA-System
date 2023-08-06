package adminController

import (
	"QA-system/apiException"
	"QA-system/controller/req"
	"QA-system/service/listService"
	"QA-system/service/questionService"
	"QA-system/utility"
	"github.com/gin-gonic/gin"
	"log"
)

func UpdateStatus(c *gin.Context) {
	var postForm req.ReqUpdateStatusData
	err := c.ShouldBindJSON(&postForm)
	if err != nil {
		_ = c.AbortWithError(200, apiException.ParamError)
		return
	}
	listService.UpdateStatus(postForm)
	utility.JsonSuccessResponse(c, nil)
}

func UpdateList(c *gin.Context) {
	var postForm req.ReqUpdateListData
	err := c.ShouldBindJSON(&postForm)
	log.Println(postForm)
	if err != nil {
		_ = c.AbortWithError(200, apiException.ParamError)
		return
	}
	if err = questionService.UpdateQuestion(postForm.ID, postForm.List); err != nil {
		_ = c.AbortWithError(200, apiException.ServerError)
		return
	}
	utility.JsonSuccessResponse(c, nil)
}
