package utils

import (
	"errors"

	"QA-System/internal/global/config"
	WeJHSDK "github.com/zjutjh/WeJH-SDK"
)

var encryptKey string

// Init 读入 AES 密钥配置
func Init() error {
	encryptKey = global.Config.GetString("aes.key")
	if len(encryptKey) != 16 && len(encryptKey) != 24 && len(encryptKey) != 32 {
		return errors.New("AES 密钥长度必须为 16、24 或 32 字节")
	}
	return nil
}

// AesEncrypt AES加密
func AesEncrypt(orig string) string {
	return WeJHSDK.AesEncrypt(orig, encryptKey)
}

// AesDecrypt AES解密
func AesDecrypt(cryted string) string {
	return WeJHSDK.AesDecrypt(cryted, encryptKey)
}
