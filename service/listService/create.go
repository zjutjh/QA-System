package listService

import (
	"QA-system/config/database"
	"QA-system/controller/req"
	"QA-system/model"
	"QA-system/service/questionService"
)

func CreateList(data req.CreateListData) error {
	list := model.List{
		Title:  data.Title,
		Num:    0,
		Draft:  true,
		Public: false,
	}
	result := database.DB.Create(&list)
	if result.Error != nil {
		return result.Error
	}
	return questionService.Create(list.ID, data.Questions)
}
