package code

import (
	"net/http"

	"QA-System/internal/pkg/log"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Error 表示自定义错误，包括状态码、消息和日志级别。
type Error struct {
	Code  int
	Msg   string
	Level log.Level
}

// Error 表示自定义的错误类型
var (
	ServerError           = NewError(200500, log.LevelError, "系统异常，请稍后重试!")
	ParamError            = NewError(200501, log.LevelInfo, "参数错误")
	UserNotFind           = NewError(200502, log.LevelInfo, "该用户不存在")
	NotLogin              = NewError(200503, log.LevelInfo, "未登录")
	NoThatPasswordOrWrong = NewError(200504, log.LevelInfo, "密码错误")
	HttpTimeout           = NewError(200505, log.LevelInfo, "请求超时，请稍后重试!")
	RequestError          = NewError(200506, log.LevelInfo, "系统异常，请稍后重试!")
	StatusOpenError       = NewError(200507, log.LevelInfo, "问卷状态不为未发布，请下架后重试！")
	SurveyNumError        = NewError(200508, log.LevelInfo, "问卷已有填写记录，无法修改！")
	TimeBeyondError       = NewError(200509, log.LevelInfo, "问卷不在填写时间范围内，无法填写！")
	SurveyError           = NewError(200510, log.LevelInfo, "问卷设置内容不符合规范！")
	UniqueError           = NewError(200511, log.LevelInfo, "唯一问题的填写内容重复，请重新填写！")
	UserExist             = NewError(200512, log.LevelInfo, "该用户已存在")
	PictureError          = NewError(200513, log.LevelInfo, "仅允许上传图片文件")
	FileSizeError         = NewError(200514, log.LevelInfo, "文件大小超出限制")
	NotSuperAdmin         = NewError(200513, log.LevelInfo, "很抱歉，您暂无权限注册账号")
	NoPermission          = NewError(200514, log.LevelInfo, "很抱歉，您暂无权限操作")
	SurveyNotExist        = NewError(200515, log.LevelInfo, "问卷不存在")
	PermissionExist       = NewError(200516, log.LevelInfo, "该用户已有权限，请勿重复操作！")
	PermissionBelong      = NewError(200517, log.LevelInfo, "问卷为该用户所有，无需操作！")
	PermissionNotExist    = NewError(200518, log.LevelInfo, "该用户无权限，请勿操作！")
	SurveyIncomplete      = NewError(200519, log.LevelInfo, "问卷未填写完整，请重新检查！")
	SurveyContentRepeat   = NewError(200520, log.LevelInfo, "问卷问题或选项重复，请重新填写！")
	NewPasswordSame       = NewError(200521, log.LevelInfo, "新密码与旧密码相同")
	SurveyNotOpen         = NewError(200522, log.LevelInfo, "问卷未开放")
	UserNotFound          = NewError(200523, log.LevelInfo, "用户不存在")
	VoteLimitError        = NewError(200524, log.LevelInfo, "单日投票次数已达上限")
	OptionNumError        = NewError(200525, log.LevelInfo, "选项设置不符合要求")
	PageBeyondError       = NewError(200526, log.LevelInfo, "页码超出范围,请重新选择")
	SurveyTypeError       = NewError(200527, log.LevelInfo, "问卷类型错误")
	OauthTimeError        = NewError(200528, log.LevelInfo, "统一登录在夜晚不可用，请在白天尝试")
	StatusRepeatError     = NewError(200529, log.LevelInfo, "问卷状态重复，请重新选择")
	AnswerSheetNotExist   = NewError(200530, log.LevelInfo, "答卷不存在,请重新选择")
	VoteSumLimitError     = NewError(200531, log.LevelInfo, "总投票次数已达上限")
	NotUnderGraduateError = NewError(200532, log.LevelInfo, "当前问卷仅允许本科生提交")
	NotFound              = NewError(200404, log.LevelInfo, http.StatusText(http.StatusNotFound))
)

// Error 方法实现了 error 接口，返回错误的消息内容
func (e *Error) Error() string {
	return e.Msg
}

// NewError 创建并返回一个新的自定义错误实例
func NewError(code int, level log.Level, msg string) *Error {
	return &Error{
		Code:  code,
		Msg:   msg,
		Level: level,
	}
}

// AbortWithException 用于返回自定义错误信息
func AbortWithException(c *gin.Context, apiError *Error, err error) {
	logError(c, apiError, err)
	_ = c.AbortWithError(200, apiError) //nolint:errcheck
}

// logError 记录错误日志
func logError(c *gin.Context, apiErr *Error, err error) {
	// 构建日志字段
	logFields := []zap.Field{
		zap.Int("error_code", apiErr.Code),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.String("ip", c.ClientIP()),
		zap.Error(err), // 记录原始错误信息
	}
	log.GetLogFunc(apiErr.Level)(apiErr.Msg, logFields...)
}
