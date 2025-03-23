package utils

import (
	"errors"
	"time"

	global "QA-System/internal/global/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/zjutjh/WeJH-SDK/oauth"
)

var (
	key string
	t   *jwt.Token
)

// NewJWT 生成 JWT
func NewJWT(name, college, stuId, userType, userTypeDesc, gender string) string {
	key = global.Config.GetString("jwt.secret")
	duration := time.Hour * 24 * 7
	expirationTime := time.Now().Add(duration).Unix() // 设置过期时间
	t = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"name":         name,
		"college":      college,
		"stuId":        stuId,
		"userType":     userType,
		"userTypeDesc": userTypeDesc,
		"gender":       gender,
		"exp":          expirationTime,
	})
	s, err := t.SignedString([]byte(key))
	if err != nil {
		return ""
	}
	return s
}

// ParseJWT 解析 JWT
func ParseJWT(token string) (oauth.UserInfo, error) {
	key = global.Config.GetString("jwt.secret")
	t, err := jwt.Parse(token, func(_ *jwt.Token) (any, error) {
		return []byte(key), nil
	})
	if err != nil {
		return oauth.UserInfo{}, err
	}
	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok || !t.Valid { // 检查令牌是否有效
		return oauth.UserInfo{}, errors.New("invalid token")
	}

	// 验证 exp 是否有效
	if exp, ok := claims["exp"].(float64); ok {
		if time.Unix(int64(exp), 0).Before(time.Now()) {
			return oauth.UserInfo{}, errors.New("token expired")
		}
	}

	keys := []string{"name", "college", "stuId", "userType", "userTypeDesc", "gender"}
	values := make(map[string]string)
	for _, k := range keys {
		v, ok := claims[k].(string)
		if !ok {
			return oauth.UserInfo{}, errors.New("invalid token")
		}
		values[k] = v
	}
	userInfo := oauth.UserInfo{
		Name:         values["name"],
		College:      values["college"],
		UserType:     values["userType"],
		UserTypeDesc: values["userTypeDesc"],
		Gender:       values["gender"],
		StudentID:    values["stuId"],
	}
	return userInfo, nil
}
