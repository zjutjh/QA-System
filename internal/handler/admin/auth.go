package admin

import (
	"QA-System/internal/models"
	"QA-System/internal/pkg/code"
	"QA-System/internal/pkg/utils"
	"QA-System/internal/service"
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LoginData struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// 登录
func Login(c *gin.Context) {
	var data LoginData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("登录失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ParamError)
		return
	}
	//判断密码是否正确
	user, err := service.GetAdminByUsername(data.Username)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("用户信息获取失败的原因: " + err.Error()), Type: gin.ErrorTypeAny})
		if err == gorm.ErrRecordNotFound {
			utils.JsonErrorResponse(c, code.UserNotFind)
			return
		} else {
			utils.JsonErrorResponse(c, code.ServerError)
			return
		}
	}
	if user.Password != data.Password {
		c.Error(&gin.Error{Err: errors.New("密码错误"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NoThatPasswordOrWrong)
		return
	}
	//设置session
	err = service.SetUserSession(c, user)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("设置session失败的原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}

	utils.JsonSuccessResponse(c, nil)
}

type RegisterData struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Key      string `json:"key" binding:"required"`
}

// 注册
func Register(c *gin.Context) {
	var data RegisterData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("注册失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ParamError)
		return
	}
	//判断是否有权限
	adminKey := service.GetConfigKey()
	if adminKey != data.Key {
		c.Error(&gin.Error{Err: errors.New(data.Username + "没有权限"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NotSuperAdmin)
		return
	}
	//判断用户是否存在
	err = service.IsAdminExist(data.Username)
	if err == nil {
		c.Error(&gin.Error{Err: errors.New(data.Username + "用户已存在"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.UserExist)
		return
	}
	//创建用户
	err = service.CreateAdmin(models.User{
		Username:  data.Username,
		Password:  data.Password,
		AdminType: 1,
	})
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("创建用户失败的原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}

	utils.JsonSuccessResponse(c, nil)
}

type UpdatePasswordData struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// 修改密码
func UpdatePassword(c *gin.Context) {
	var data UpdatePasswordData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("重置密码失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ParamError)
		return
	}
	//判断用户是否存在
	user, err := service.GetUserSession(c)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取用户缓存信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NotLogin)
		return
	}
	// 判断旧密码是否正确
	if user.Password != data.OldPassword {
		c.Error(&gin.Error{Err: errors.New("旧密码错误"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NoThatPasswordOrWrong)
		return
	}
	// 判断新密码是否与旧密码相同
	if user.Password == data.NewPassword {
		c.Error(&gin.Error{Err: errors.New("新密码与旧密码相同"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NewPasswordSame)
		return
	}
	//修改密码
	err = service.UpdateAdminPassword(user.ID, data.NewPassword)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("修改密码失败的原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	utils.JsonSuccessResponse(c, nil)
}

type ResetPasswordData struct {
	UserName string `json:"username" binding:"required"`
}

// 重置密码
func ResetPassword(c *gin.Context) {
	var data ResetPasswordData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("重置密码失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ParamError)
		return
	}
	//鉴权
	admin, err := service.GetUserSession(c)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取用户缓存信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NotLogin)
		return
	}
	if admin.AdminType != 2 {
		c.Error(&gin.Error{Err: errors.New(admin.Username + "没有权限"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.NoPermission)
		return
	}
	//判断用户是否存在
	user, err := service.GetAdminByUsername(data.UserName)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取用户信息失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	//重置密码
	err = service.UpdateAdminPassword(user.ID, "jhwl")
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("重置密码失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	utils.JsonSuccessResponse(c, nil)
}
