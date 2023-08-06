package model

type Ans struct {
	ID         uint   `json:"-",gorm:"primary_key"`
	ListID     uint   `json:"-"`
	QuestionID uint   `json:"qid"`
	Content    string `json:"content"`
	NumID      int    `json:"-"`
}
