package model

type List struct {
	Title  string `json:"title"`
	ID     uint   `json:"id",gorm:"primary_key"`
	Num    uint   `json:"-"`
	Draft  bool   `json:"draft"`
	Public bool   `json:"public"`
}
