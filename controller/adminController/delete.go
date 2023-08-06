package adminController

import (
	"QA-system/apiException"
	"QA-system/controller/req"
	"QA-system/service/listService"
	"QA-system/utility"
	"github.com/gin-gonic/gin"
)

func Delete(c *gin.Context) {
	var postForm req.ReqDeleteData
	err := c.ShouldBindJSON(&postForm)
	if err != nil {
		_ = c.AbortWithError(200, apiException.ParamError)
		return
	}
	if err = listService.Delete(postForm.ID); err != nil {
		_ = c.AbortWithError(500, apiException.ServerError)
		return
	}
	utility.JsonSuccessResponse(c, nil)
}
