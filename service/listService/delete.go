package listService

import (
	"QA-system/config/database"
	"QA-system/model"
	"QA-system/service/questionService"
)

func Delete(id uint) error {
	list, _ := GetList(id)
	database.DB.Where(model.List{}).Delete(&list)
	return questionService.DeleteQuestionByListID(id)
}
