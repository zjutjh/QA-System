package listService

import (
	"QA-system/config/database"
	"QA-system/controller/res"
	"QA-system/model"
	"QA-system/service/questionService"
)

func GetListByID(id int) (*res.ResListData, error) {
	var data res.ResListData
	var err error
	if err = database.DB.Where(model.List{ID: uint(id)}).Find(&data.List).Error; err != nil {
		return nil, err
	}
	if data.Question, err = questionService.GetQuestion(id); err != nil {
		return nil, err
	}
	return &data, nil
}

func GetList(id uint) (*model.List, error) {
	var data model.List
	if err := database.DB.Where(model.List{ID: uint(id)}).Find(&data).Error; err != nil {
		return nil, err
	}
	return &data, nil
}

func GetAll() (*[]model.List, error) {
	var data []model.List
	if err := database.DB.Find(&data).Error; err != nil {
		return nil, err
	}
	return &data, nil
}

func GetAllPublic() (*[]model.List, error) {
	var data []model.List
	if err := database.DB.Where(&model.List{
		Public: true,
		Draft:  false,
	}).Find(&data).Error; err != nil {
		return nil, err
	}
	return &data, nil
}
