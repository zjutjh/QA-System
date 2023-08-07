package model

type Ans struct {
	ID         uint   `json:"-",gorm:"primary_key"`
	ListID     uint   `json:"-"`
	QuestionID uint   `json:"qid"`
	Content    string `json:"content"`
	NumID      int    `json:"-"`
}

type AnsSort []Ans

func (a AnsSort) Len() int {
	return len(a)
}
func (a AnsSort) Less(i, j int) bool {
	return a[i].QuestionID < a[j].QuestionID
}

func (a AnsSort) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
