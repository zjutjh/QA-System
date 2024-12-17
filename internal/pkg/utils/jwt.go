package utils

import (
	"errors"
	"time"

	global "QA-System/internal/global/config"
	"github.com/golang-jwt/jwt/v5"
)

var (
	key string
	t   *jwt.Token
)

// NewJWT 生成 JWT
func NewJWT(stuId string) string {
	key = global.Config.GetString("jwt.secret")
	duration := time.Hour * 24 * 7
	expirationTime := time.Now().Add(duration).Unix() // 设置过期时间
	t = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"stuId": stuId,
		"exp":   expirationTime,
	})
	s, err := t.SignedString([]byte(key))
	if err != nil {
		return ""
	}
	return s
}

// ParseJWT 解析 JWT
func ParseJWT(token string) (string, error) {
	key = global.Config.GetString("jwt.secret")
	t, err := jwt.Parse(token, func(_ *jwt.Token) (any, error) {
		return []byte(key), nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok || !t.Valid { // 检查令牌是否有效
		return "", errors.New("invalid token")
	}

	// 验证 exp 是否有效
	if exp, ok := claims["exp"].(float64); ok {
		if time.Unix(int64(exp), 0).Before(time.Now()) {
			return "", errors.New("token expired")
		}
	}

	stuId, ok := claims["stuId"].(string)
	if !ok {
		return "", errors.New("invalid token")
	}
	return stuId, nil
}
