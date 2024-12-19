package service

import (
	"QA-System/internal/pkg/api/userCenterApi"
	"QA-System/internal/pkg/code"
	"QA-System/internal/pkg/request"
	"net/url"
)

// UserCenterResponse 用户中心响应结构体
type UserCenterResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

// FetchHandleOfPost 向用户中心发送 POST 请求
func FetchHandleOfPost(form map[string]any, url string) (*UserCenterResponse, error) {
	client := request.NewUnSafe()
	var rc UserCenterResponse

	// 发送 POST 请求并自动解析 JSON 响应
	resp, err := client.Request().
		SetHeader("Content-Type", "application/json").
		SetBody(form).
		SetResult(&rc).
		Post(userCenterApi.UserCenterHost + url)

	// 检查请求错误
	if err != nil || resp.IsError() {
		return nil, code.RequestError
	}

	// 返回解析后的响应
	return &rc, nil
}

// 统一登录验证
func Oauth(sid, password string) error {
	loginUrl, err := url.Parse(userCenterApi.OAuth)
	if err != nil {
		return err
	}
	urlPath := loginUrl.String()
	regMap := map[string]any{
		"stu_id":   sid,
		"password": password,
	}
	resp, err := FetchHandleOfPost(regMap, urlPath)
	if err != nil {
		return code.RequestError
	}

	// 使用 handleLoginErrors 函数处理响应码
	return handleLoginErrors(resp.Code)
}

// handleRegErrors 根据响应码处理不同的错误
func handleLoginErrors(num int) error {
	switch num {
	case 404:
		return code.UserNotFound
	case 409:
		return code.NoThatPasswordOrWrong
	case 408:
		return code.HttpTimeout
	case 410:
		return code.ServerError
	case 507:
		return code.OauthTimeError
	case 200:
		return nil
	default:
		return code.ServerError
	}
}
