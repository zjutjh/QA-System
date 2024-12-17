package service

import (
	"errors"

	"QA-System/internal/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// SetUserSession 设置用户会话
func SetUserSession(c *gin.Context, user *models.User) error {
	webSession := sessions.Default(c)
	webSession.Options(sessions.Options{
		MaxAge:   3600 * 24 * 7,
		Path:     "/",
		HttpOnly: true,
	})
	webSession.Set("id", user.ID)
	return webSession.Save()
}

// GetUserSession 获取用户会话
func GetUserSession(c *gin.Context) (*models.User, error) {
	webSession := sessions.Default(c)
	id := webSession.Get("id")
	if id == nil {
		return nil, errors.New("")
	}
	uid, ok := id.(int)
	if !ok {
		return nil, errors.New("")
	}
	user, err := GetAdminByID(uid)
	if user == nil || err != nil {
		err = ClearUserSession(c)
		if err != nil {
			return nil, err
		}
		return nil, errors.New("")
	}
	return user, nil
}

// UpdateUserSession 更新用户会话
func UpdateUserSession(c *gin.Context) (*models.User, error) {
	user, err := GetUserSession(c)
	if err != nil {
		return nil, err
	}
	err = SetUserSession(c, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// CheckUserSession 检查用户会话
func CheckUserSession(c *gin.Context) bool {
	webSession := sessions.Default(c)
	id := webSession.Get("id")
	return id != nil
}

// ClearUserSession 清除用户会话
func ClearUserSession(c *gin.Context) error {
	webSession := sessions.Default(c)
	webSession.Delete("id")
	err := webSession.Save()
	if err != nil {
		return err
	}
	return nil
}
