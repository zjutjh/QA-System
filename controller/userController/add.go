package userController

import (
	"QA-system/apiException"
	"QA-system/controller/req"
	"QA-system/service/ansService"
	"QA-system/service/listService"
	"QA-system/utility"
	"github.com/gin-gonic/gin"
)

func AddAns(c *gin.Context) {
	var postForm req.ReqAnsData
	err := c.ShouldBindJSON(&postForm)
	if err != nil {
		_ = c.AbortWithError(200, apiException.ParamError)
		return
	}
	list, err := listService.GetList(postForm.ID)
	if list.Public == false || err != nil {
		_ = c.AbortWithError(200, apiException.QAListNotExit)
		return
	}
	num, err := listService.AddNum(postForm.ID)
	ansService.CreateAns(num, postForm.ID, postForm.List)
	utility.JsonSuccessResponse(c, nil)
}
