package model

// BaseConfig 问卷基本配置（请求 DTO）
type BaseConfig struct {
	StartTime     string `json:"start_time" binding:"datetime=2006-01-02T15:04:05+08:00"`
	EndTime       string `json:"end_time" binding:"datetime=2006-01-02T15:04:05+08:00"`
	DailyLimit    uint   `json:"day_limit"`      // 问卷每日填写限制
	SumLimit      uint   `json:"sum_limit"`      // 问卷总填写次数限制
	Verify        bool   `json:"verify"`         // 问卷是否需要统一验证
	UndergradOnly bool   `json:"undergrad_only"` // 是否只限制本科生作答
	NeedNotify    bool   `json:"need_notify"`    // 问卷在收到回复时是否需要提醒
}

// QuestionConfig 问卷问题配置（请求 DTO）
type QuestionConfig struct {
	Desc         string         `json:"desc"`
	Title        string         `json:"title"`
	QuestionList []QuestionList `json:"question_list"`
}

// QuestionList 问题列表项（请求 DTO）
type QuestionList struct {
	SerialNum       int             `json:"serial_num"`   // 题目序号
	Subject         string          `json:"subject"`      // 问题
	Description     string          `json:"description"`  // 问题描述
	Img             string          `json:"img"`          // 图片
	QuestionSetting QuestionSetting `json:"ques_setting"` // 问题设置
	Options         []OptionItem    `json:"options"`      // 选项
}

// QuestionSetting 问题设置（请求 DTO）
type QuestionSetting struct {
	Required      bool         `json:"required"`                                           // 是否必填
	Unique        bool         `json:"unique"`                                             // 是否唯一
	OtherOption   bool         `json:"other_option"`                                       // 是否有其他选项
	QuestionType  int          `json:"question_type" binding:"required,oneof=1 2 3 4 5 6"` // 问题类型
	Reg           string       `json:"reg"`                                                // 正则表达式
	Options       []OptionItem `json:"options"`                                            // 选项
	MaximumOption uint         `json:"maximum_option"`                                     // 多选最多选项数 0为不限制
	MinimumOption uint         `json:"minimum_option"`                                     // 多选最少选项数 0为不限制
}

// OptionItem 选项数据项（请求 DTO，区别于 model.Option 的 ORM 模型）
type OptionItem struct {
	SerialNum   int    `json:"serial_num"`  // 选项序号
	Content     string `json:"content"`     // 选项内容
	Description string `json:"description"` // 选项描述
	Img         string `json:"img"`         // 图片
}

// QuestionSubmit 用户提交的问题答案（请求 DTO）
type QuestionSubmit struct {
	QuestionID int    `json:"question_id" binding:"required"`
	Answer     string `json:"answer"`
}
