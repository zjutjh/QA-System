package adminController

import (
	"QA-system/apiException"
	"QA-system/service/ansService"
	"QA-system/service/listService"
	"QA-system/service/questionService"
	"QA-system/utility"
	"github.com/gin-gonic/gin"
	"net/url"
	"os"
	"strconv"
)

func ExportExcel(c *gin.Context) {
	tid, _ := url.QueryUnescape(c.Param("list_id"))
	id, err := strconv.Atoi(tid)
	if err != nil {
		_ = c.AbortWithError(500, apiException.ParamError)
		return
	}
	list, _ := listService.GetList(uint(id))
	list2, _ := questionService.GetQuestion(id)
	var titleList []string
	var dataList [][]string
	ansList, err := ansService.GetAns(id)
	for _, index := range ansList {
		var temp []string
		for _, j := range index {
			temp = append(temp, j.Content)
		}
		dataList = append(dataList, temp)
	}
	for _, index := range list2 {
		titleList = append(titleList, index.Text)
	}
	filename := utility.DataToExcel(titleList, dataList, list.Title)
	c.FileAttachment(filename, filename)
	os.Remove(filename)
	return
}
