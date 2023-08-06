package questionService

import (
	"QA-system/config/database"
	"QA-system/model"
)

func GetQuestion(id int) ([]model.Question, error) {
	var list []model.Question
	if result := database.DB.Where(model.Question{ListID: uint(id)}).Order("id asc").Find(&list); result.Error != nil {
		return nil, result.Error
	}
	return list, nil
}
