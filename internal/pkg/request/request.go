package request

import (
	"QA-System/internal/pkg/log"
	"crypto/tls"
	"encoding/json"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"

	"time"
)

// Client 包装 Resty 客户端
type Client struct {
	*resty.Client
}

// New 初始化一个 Resty 客户端
func New() Client {
	s := Client{
		Client: resty.New().
			SetTimeout(5 * time.Second).
			SetJSONMarshaler(jsoniter.ConfigCompatibleWithStandardLibrary.Marshal).
			SetJSONUnmarshaler(jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal).
			// 配置重试机制
			SetRetryCount(3).                     // 设置重试次数
			SetRetryMaxWaitTime(2 * time.Second), // 最大重试等待时间,
	}
	// 添加重试条件：仅对特定的 HTTP 状态码或错误类型重试
	s.Client.AddRetryCondition(func(r *resty.Response, err error) bool {
		// 如果发生网络错误，或者返回的 HTTP 状态码是 5xx，执行重试
		if err != nil || (r.StatusCode() >= 500 && r.StatusCode() < 600) {
			return true
		}
		return false
	})
	// 添加重试条件
	s.Client.AddRetryCondition(func(r *resty.Response, err error) bool {
		if err != nil {
			// 网络错误时重试
			log.Logger.Error("Network error: %v. Retrying..." + err.Error())
			return true
		}

		// 如果 HTTP 状态码不是 200，通常不需要解析 JSON，直接返回
		if r.StatusCode() != 200 {
			log.Logger.Error("Non-200 status code: %d. Retrying...", zap.Int("status_code", r.StatusCode()))
			return true
		}

		// 解析响应 JSON 数据
		var resp struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
			Data any    `json:"data"`
		}

		if unmarshalErr := json.Unmarshal(r.Body(), &resp); unmarshalErr != nil {
			// JSON 解析失败，可能是服务器返回非预期数据
			log.Logger.Error("Failed to parse response JSON: %v. Retrying...", zap.Error(unmarshalErr))
			return true
		}

		// 根据业务逻辑判断是否需要重试
		switch resp.Code {
		case 200:
			return false
		case 409:
			return false
		default:
			return true
		}
	})

	// 利用中间件实现请求日志
	s.OnAfterResponse(RestyLogMiddleware)

	return s

}

// NewUnSafe 初始化一个 Resty 客户端并跳过 TLS 证书验证
func NewUnSafe() Client {
	s := New()
	s.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	return s
}

// Request 获取一个新的请求实例
func (s Client) Request() *resty.Request {
	return s.R().EnableTrace()
}

// RestyLogMiddleware Resty日志中间件
func RestyLogMiddleware(_ *resty.Client, resp *resty.Response) error {
	if resp.IsError() {
		method := resp.Request.Method
		url := resp.Request.URL
		log.Logger.Error("请求出现错误", zap.String("method", method), zap.String("url", url), zap.Int64("time_spent(ms)", resp.Time().Milliseconds()))
	}
	return nil
}
