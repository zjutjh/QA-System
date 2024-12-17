package admin

import (
	"errors"

	"QA-System/internal/models"
	"QA-System/internal/pkg/code"
	"QA-System/internal/pkg/utils"
	"QA-System/internal/service"
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
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	//判断密码是否正确
	user, err := service.GetAdminByUsername(data.Username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			code.AbortWithException(c, code.UserNotFind, err)
			return
		} else {
			code.AbortWithException(c, code.ServerError, err)
			return
		}
	}
	if user.Password != data.Password {
		code.AbortWithException(c, code.NoThatPasswordOrWrong, errors.New("密码错误"))
		return
	}
	//设置session
	err = service.SetUserSession(c, user)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
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
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	//判断是否有权限
	adminKey := service.GetConfigKey()
	if adminKey != data.Key {
		code.AbortWithException(c, code.NotSuperAdmin, errors.New(data.Username+"没有权限"))
		return
	}
	//判断用户是否存在
	err = service.IsAdminExist(data.Username)
	if err == nil {
		code.AbortWithException(c, code.UserExist, errors.New(data.Username+"用户已存在"))
		return
	}
	//创建用户
	err = service.CreateAdmin(models.User{
		Username:  data.Username,
		Password:  data.Password,
		AdminType: 1,
	})
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
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
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	//判断用户是否存在
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
	//修改密码
	err = service.UpdateAdminPassword(user.ID, data.NewPassword)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
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
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	//鉴权
	admin, err := service.GetUserSession(c)
	if err != nil {
		code.AbortWithException(c, code.NotLogin, err)
		return
	}
	if admin.AdminType != 2 {
		code.AbortWithException(c, code.NoPermission, errors.New(admin.Username+"没有权限"))
		return
	}
	//判断用户是否存在
	user, err := service.GetAdminByUsername(data.UserName)
	if err != nil {
		code.AbortWithException(c, code.UserNotFind, err)
		return
	}
	//重置密码
	err = service.UpdateAdminPassword(user.ID, "jhwl")
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	utils.JsonSuccessResponse(c, nil)
}
