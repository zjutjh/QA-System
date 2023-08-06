package questionService

import "QA-system/model"

func UpdateQuestion(id uint, list []model.Question) error {
	err := DeleteQuestionByListID(id)
	if err != nil {
		return err
	}
	return Create(id, list)
}
