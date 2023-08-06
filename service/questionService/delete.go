package questionService

import (
	"QA-system/config/database"
	"QA-system/model"
)

func DeleteQuestionByListID(id uint) error {
	err := database.DB.Delete(&model.Question{}, "list_id = ?", id).Error
	return err
}
