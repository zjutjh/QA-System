package req

import "QA-system/model"

type CreateListData struct {
	Title     string           `json:"title"`
	Questions []model.Question `json:"list"`
}

type ReqAnsData struct {
	ID   uint        `json:"id"`
	List []model.Ans `json:"list"`
}

type ReqUpdateStatusData struct {
	ID     uint `json:"id"`
	Draft  bool `json:"draft"`
	Public bool `json:"public"`
}

type ReqUpdateListData struct {
	ID   uint             `json:"id"`
	List []model.Question `json:"list"`
}

type ReqDeleteData struct {
	ID uint `json:"id"`
}
