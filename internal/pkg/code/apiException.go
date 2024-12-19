package code

import "net/http"

type Error struct {
	StatusCode int    `json:"-"`
	Code       int    `json:"code"`
	Msg        string `json:"msg"`
}

func (e Error) Error() string {
	return e.Msg
}

var (
	ServerError           = NewError(http.StatusInternalServerError, 200500, "系统异常，请稍后重试!")
	ParamError            = NewError(http.StatusInternalServerError, 200501, "参数错误")
	UserNotFind           = NewError(http.StatusInternalServerError, 200502, "该用户不存在")
	NotLogin              = NewError(http.StatusInternalServerError, 200503, "未登录")
	NoThatPasswordOrWrong = NewError(http.StatusInternalServerError, 200504, "密码错误")
	HttpTimeout           = NewError(http.StatusInternalServerError, 200505, "请求超时，请稍后重试!")
	RequestError          = NewError(http.StatusInternalServerError, 200506, "系统异常，请稍后重试!")
	StatusOpenError       = NewError(http.StatusInternalServerError, 200507, "问卷状态不为未发布，请下架后重试！")
	SurveyNumError        = NewError(http.StatusInternalServerError, 200508, "问卷已有填写记录，无法修改！")
	TimeBeyondError       = NewError(http.StatusInternalServerError, 200509, "问卷不在填写时间范围内，无法填写！")
	SurveyError           = NewError(http.StatusInternalServerError, 200510, "问卷设置内容不符合规范！")
	UniqueError           = NewError(http.StatusInternalServerError, 200511, "唯一问题的填写内容重复，请重新填写！")
	UserExist             = NewError(http.StatusInternalServerError, 200512, "该用户已存在")
	PictureError          = NewError(http.StatusInternalServerError, 200513, "仅允许上传图片文件")
	PictureSizeError      = NewError(http.StatusInternalServerError, 200514, "图片大小超出限制")
	NotSuperAdmin         = NewError(http.StatusInternalServerError, 200513, "很抱歉，您暂无权限注册账号")
	NoPermission          = NewError(http.StatusInternalServerError, 200514, "很抱歉，您暂无权限操作")
	SurveyNotExist        = NewError(http.StatusInternalServerError, 200515, "问卷不存在")
	PermissionExist       = NewError(http.StatusInternalServerError, 200516, "该用户已有权限，请勿重复操作！")
	PermissionBelong      = NewError(http.StatusInternalServerError, 200517, "问卷为该用户所有，无需操作！")
	PermissionNotExist    = NewError(http.StatusInternalServerError, 200518, "该用户无权限，请勿操作！")
	SurveyIncomplete      = NewError(http.StatusInternalServerError, 200519, "问卷未填写完整，请重新检查！")
	SurveyContentRepeat   = NewError(http.StatusInternalServerError, 200520, "问卷问题或选项重复，请重新填写！")
	NewPasswordSame       = NewError(http.StatusInternalServerError, 200521, "新密码与旧密码相同")
	SurveyNotOpen         = NewError(http.StatusInternalServerError, 200522, "问卷未开放")
	UserNotFound          = NewError(http.StatusInternalServerError, 200523, "用户不存在")
	VoteLimitError        = NewError(http.StatusInternalServerError, 200524, "投票次数已达上限")
	OptionNumError        = NewError(http.StatusInternalServerError, 200525, "选项设置不符合要求")
	StuIDRedisError       = NewError(http.StatusInternalServerError, 200526, "未统一验证，请重新进入并进行统一验证")
	SurveyTypeError       = NewError(http.StatusInternalServerError, 200527, "问卷类型错误")
	OauthTimeError        = NewError(http.StatusInternalServerError, 200528, "统一登录在夜晚不可用，请在白天尝试")
	StatusRepeatError     = NewError(http.StatusInternalServerError, 200529, "问卷状态重复，请重新选择")
	NotInit               = NewError(http.StatusNotFound, 200404, http.StatusText(http.StatusNotFound))
	NotFound              = NewError(http.StatusNotFound, 200404, http.StatusText(http.StatusNotFound))
	Unknown               = NewError(http.StatusInternalServerError, 300500, "系统异常，请稍后重试!")
)

func NewError(statusCode, Code int, msg string) *Error {
	return &Error{
		StatusCode: statusCode,
		Code:       Code,
		Msg:        msg,
	}
}
