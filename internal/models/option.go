package models

type Option struct {
	ID          int    `json:"id"`
	QuestionID  int    `json:"question_id"` //问题ID
	SerialNum   int    `json:"serial_num"`  //选项序号
	Content     string `json:"content"`     //选项内容
	Description string `json:"description"` //选项描述
	Img         string `json:"img"`         //选项图片
}
