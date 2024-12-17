package models

// Type 预先信息模型
type Type struct {
	ID    int    `json:"id"`    // 类型id
	Type  string `json:"type"`  // 类型
	Value string `json:"value"` // 类型值
}
