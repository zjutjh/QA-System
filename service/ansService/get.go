package ansService

import (
	"QA-system/config/database"
	"QA-system/model"
)

func GetAns(id int) ([][]model.Ans, error) {
	var ansList []model.Ans
	database.DB.Where(model.Ans{ListID: uint(id)}).Order("num_id asc").Find(&ansList)
	ansList = append(ansList, model.Ans{
		ID:         0,
		ListID:     0,
		QuestionID: 0,
		Content:    "",
		NumID:      -1,
	})
	var data [][]model.Ans
	flag := ansList[0]
	length := 0
	for index, ans := range ansList {
		if flag.NumID != ans.NumID {
			flag = ans
			temp := ansList[index-length : index]
			data = append(data, temp)
			length = 0
		}
		length++
	}
	return data, nil
}
