package admin

import (
	"errors"

	"QA-System/internal/model"
	"QA-System/internal/pkg/code"
	"QA-System/internal/pkg/utils"
	"QA-System/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type loginData struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 登录
func Login(c *gin.Context) {
	var data loginData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	// 判断密码是否正确
	user, err := service.GetAdminByUsername(data.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			code.AbortWithException(c, code.UserNotFind, err)
			return
		}
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	if user.Password != data.Password {
		code.AbortWithException(c, code.NoThatPasswordOrWrong, errors.New("密码错误"))
		return
	}
	// 设置session
	err = service.SetUserSession(c, user)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}

	utils.JsonSuccessResponse(c, nil)
}

type registerData struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	Key         string `json:"key" binding:"required"`
	NotifyEmail string `json:"notify_email" binding:"email"`
}

// Register 注册
func Register(c *gin.Context) {
	var data registerData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	// 判断是否有权限
	adminKey := service.GetConfigKey()
	if adminKey != data.Key {
		code.AbortWithException(c, code.NotSuperAdmin, errors.New(data.Username+"没有权限"))
		return
	}
	// 判断用户是否存在
	err = service.IsAdminExist(data.Username)
	if err == nil {
		code.AbortWithException(c, code.UserExist, errors.New(data.Username+"用户已存在"))
		return
	}
	// 创建用户
	err = service.CreateAdmin(model.User{
		Username:    data.Username,
		Password:    data.Password,
		AdminType:   1,
		NotifyEmail: data.NotifyEmail,
	})
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}

	utils.JsonSuccessResponse(c, nil)
}

type updatePasswordData struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// UpdatePassword 修改密码
func UpdatePassword(c *gin.Context) {
	var data updatePasswordData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	// 判断用户是否存在
	user, err := service.GetUserSession(c)
	if err != nil {
		code.AbortWithException(c, code.NotLogin, err)
		return
	}
	// 判断旧密码是否正确
	if user.Password != data.OldPassword {
		code.AbortWithException(c, code.NoThatPasswordOrWrong, errors.New("旧密码错误"))
		return
	}
	// 判断新密码是否与旧密码相同
	if user.Password == data.NewPassword {
		code.AbortWithException(c, code.NewPasswordSame, errors.New("新密码与旧密码相同"))
		return
	}
	// 修改密码
	err = service.UpdateAdminPassword(user.ID, data.NewPassword)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	utils.JsonSuccessResponse(c, nil)
}

type resetPasswordData struct {
	UserName string `json:"username" binding:"required"`
}

// ResetPassword 重置密码
func ResetPassword(c *gin.Context) {
	var data resetPasswordData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	// 鉴权
	admin, err := service.GetUserSession(c)
	if err != nil {
		code.AbortWithException(c, code.NotLogin, err)
		return
	}
	if admin.AdminType != 2 {
		code.AbortWithException(c, code.NoPermission, errors.New(admin.Username+"没有权限"))
		return
	}
	// 判断用户是否存在
	user, err := service.GetAdminByUsername(data.UserName)
	if err != nil {
		code.AbortWithException(c, code.UserNotFind, err)
		return
	}
	// 重置密码
	err = service.UpdateAdminPassword(user.ID, "jhwl")
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	utils.JsonSuccessResponse(c, nil)
}

type updateEmail struct {
	NewEmail string `json:"new_email" binding:"required,email"`
}

// UpdateEmail 修改邮箱
func UpdateEmail(c *gin.Context) {
	var data updateEmail
	err := c.ShouldBindJSON(&data)
	if err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	// 判断用户是否存在
	user, err := service.GetUserSession(c)
	if err != nil {
		code.AbortWithException(c, code.NotLogin, err)
		return
	}
	// 修改邮箱
	err = service.UpdateAdminEmail(user.ID, data.NewEmail)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	utils.JsonSuccessResponse(c, nil)
}

// requireLogin 从 session 获取当前登录管理员，未登录时直接中断请求并返回 false。
func requireLogin(c *gin.Context) (*model.User, bool) {
	user, err := service.GetUserSession(c)
	if err != nil {
		code.AbortWithException(c, code.NotLogin, err)
		return nil, false
	}
	return user, true
}

// requireSuperAdmin 要求当前登录用户是超级管理员（AdminType == 2）。
// 未登录或权限不足时直接中断请求并返回 false。
func requireSuperAdmin(c *gin.Context) (*model.User, bool) {
	user, ok := requireLogin(c)
	if !ok {
		return nil, false
	}
	if user.AdminType != 2 {
		code.AbortWithException(c, code.NoPermission, errors.New(user.Username+"没有权限"))
		return nil, false
	}
	return user, true
}

// requireSurveyPermission 要求当前登录用户对指定问卷有操作权限。
// 权限规则：超级管理员(2) 通过；普通管理员(1) 且是问卷拥有者 通过；被授权管理用户 通过。
// 返回当前用户和问卷，失败时直接中断请求。
func requireSurveyPermission(c *gin.Context, surveyID int64) (*model.User, *model.Survey, bool) {
	user, ok := requireLogin(c)
	if !ok {
		return nil, nil, false
	}
	survey, err := service.GetSurveyByID(surveyID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return nil, nil, false
	}
	if (user.AdminType != 2) && (user.AdminType != 1 || survey.UserID != user.ID) &&
		!service.UserInManage(user.ID, survey.ID) {
		code.AbortWithException(c, code.NoPermission, errors.New(user.Username+"无权限"))
		return nil, nil, false
	}
	return user, survey, true
}
