package models

// User 用户模型
type User struct {
	ID        int    `json:"id"`         // 用户id
	Username  string `json:"username"`   // 用户名
	Password  string `json:"password"`   // 密码
	AdminType int    `json:"admin_type"` // 1:普通管理员	2:超级管理员
}
