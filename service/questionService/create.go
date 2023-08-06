package questionService

import (
	"QA-system/config/database"
	"QA-system/model"
)

func Create(id uint, list []model.Question) error {
	for _, question := range list {
		database.DB.Create(&model.Question{
			Text:    question.Text,
			Options: question.Options,
			Type:    question.Type,
			ListID:  id,
		})
	}
	return nil
}
