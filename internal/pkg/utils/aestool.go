package utils

import (
	"QA-System/internal/global/config"
	WeJHSDK "github.com/zjutjh/WeJH-SDK"
)

func AesEncrypt(orig string) string {
	key := global.Config.GetString("aes.key")
	return WeJHSDK.AesEncrypt(orig, key)
}

func AesDecrypt(cryted string) string {
	key := global.Config.GetString("aes.key")
	return WeJHSDK.AesDecrypt(cryted, key)
}
