package userCenterApi

import global "QA-System/internal/global/config"

// UserCenterHost 用户中心地址
var UserCenterHost = global.Config.GetString("user.host")

// 用户中心接口
const (
	UCRegWithoutVerify string = "api/activation/notVerify"
	UCReg              string = "api/activation"
	VerifyEmail        string = "api/verify/email"
	ReSendEmail        string = "api/email"
	Auth               string = "api/auth"
	RePass             string = "api/changePwd" //nolint
	RePassWithoutEmail string = "api/repass"
	DelAccount         string = "api/del"
	OAuth              string = "api/oauth"
)
