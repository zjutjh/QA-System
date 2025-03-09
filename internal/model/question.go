package model

// Question 问题模型
type Question struct {
	ID            int    `json:"id"`
	SurveyID      string `json:"survey_id"`      // 问卷UUID
	SerialNum     int    `json:"serial_num"`     // 题目序号
	Img           string `json:"img"`            // 图片
	Subject       string `json:"subject"`        // 题目
	Description   string `json:"description"`    // 题目描述
	Required      bool   `json:"required"`       // 是否必填
	Unique        bool   `json:"unique"`         // 是否唯一
	OtherOption   bool   `json:"other_option"`   // 是否有其他选项
	QuestionType  int    `json:"question_type"`  // 题目类型 1单选2多选3填空4简答5图片6文件
	MaximumOption uint   `json:"maximum_option"` // 多选最多所选选项数 0为不限制
	MinimumOption uint   `json:"minimum_option"` // 多选最少所选选项数 0为不限制
	Reg           string `json:"reg"`            // 正则表达式
}
