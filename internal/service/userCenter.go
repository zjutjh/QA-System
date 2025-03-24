package service

import (
	"QA-System/internal/pkg/api/userCenterApi"
	"QA-System/internal/pkg/code"
	"QA-System/internal/pkg/request"
	"github.com/zjutjh/WeJH-SDK/oauth"
)

// UserCenterResponse 用户中心响应结构体
type UserCenterResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

// FetchHandleOfPost 向用户中心发送 POST 请求
func FetchHandleOfPost(form map[string]any, webUrl string) (*UserCenterResponse, error) {
	client := request.NewUnSafe()
	var rc UserCenterResponse

	// 发送 POST 请求并自动解析 JSON 响应
	resp, err := client.Request().
		SetHeader("Content-Type", "application/json").
		SetBody(form).
		SetResult(&rc).
		Post(userCenterApi.UserCenterHost + webUrl)

	// 检查请求错误
	if err != nil || resp.IsError() {
		return nil, code.RequestError
	}

	// 返回解析后的响应
	return &rc, nil
}

// Oauth 统一登录验证
func Oauth(sid, password string) (oauth.UserInfo, error) {
	_, user, err := oauth.GetUserInfo(sid, password)
	if err != nil {
		return oauth.UserInfo{}, err
	}

	return user, err
}
