package listService

import (
	"QA-system/config/database"
	"QA-system/controller/req"
)

func AddNum(id uint) (uint, error) {
	list, err := GetList(id)
	if err != nil {
		return 0, err
	}
	list.Num++
	if err = database.DB.Save(&list).Error; err != nil {
		return 0, err
	}
	return list.Num, nil
}

func UpdateStatus(data req.ReqUpdateStatusData) error {
	list, err := GetList(data.ID)
	if err != nil {
		return err
	}
	list.Draft = data.Draft
	list.Public = data.Public
	if err = database.DB.Save(&list).Error; err != nil {
		return err
	}
	return nil
}
