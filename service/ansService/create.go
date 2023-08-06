package ansService

import (
	"QA-system/config/database"
	"QA-system/model"
)

func CreateAns(num, list_id uint, list []model.Ans) error {
	for _, ans := range list {
		database.DB.Create(&model.Ans{
			QuestionID: ans.QuestionID,
			Content:    ans.Content,
			NumID:      int(num),
			ListID:     list_id,
		})
	}
	return nil
}
