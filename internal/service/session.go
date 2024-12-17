package service

import (
	"errors"

	"QA-System/internal/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

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

func GetUserSession(c *gin.Context) (*models.User, error) {
	webSession := sessions.Default(c)
	id := webSession.Get("id")
	if id == nil {
		return nil, errors.New("")
	}
	user, _ := GetAdminByID(id.(int))
	if user == nil {
		ClearUserSession(c)
		return nil, errors.New("")
	}
	return user, nil
}

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

func CheckUserSession(c *gin.Context) bool {
	webSession := sessions.Default(c)
	id := webSession.Get("id")
	return id != nil
}

func ClearUserSession(c *gin.Context) {
	webSession := sessions.Default(c)
	webSession.Delete("id")
	webSession.Save()
}
