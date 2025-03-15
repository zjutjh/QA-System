package model

import "time"

// Survey 问卷模型
type Survey struct {
	UUID       string    `json:"uuid" gorm:"primaryKey"` // 问卷uuid
	UserID     int       `json:"user_id"`                // 用户id
	Title      string    `json:"title"`                  // 问卷标题
	Desc       string    `json:"desc"`                   // 问卷描述
	StartTime  time.Time `json:"start_time"`             // 开始时间
	Deadline   time.Time `json:"deadline"`               // 截止时间
	CreatedAt  time.Time `json:"created_at"`             // 创建时间
	Status     int       `json:"status"`                 // 问卷状态  1:未发布 2:已发布 3:已截止
	DailyLimit uint      `json:"day_limit"`              // 问卷每日填写限制
	Verify     bool      `json:"verify"`                 // 问卷是否需要统一验证
	Type       uint      `json:"type"`                   // 问卷类型 0:调研 1:投票
	Num        int       `json:"num"`                    // 问卷填写数量
}

// SurveyResp 问卷响应模型
// 此模型不存入数据库，仅用于返回给前端
type SurveyResp struct {
	UUID       string `json:"uuid"`
	Title      string `json:"title"`
	Status     int    `json:"status"`
	SurveyType uint   `json:"survey_type"`
	Num        int    `json:"num"`
}
