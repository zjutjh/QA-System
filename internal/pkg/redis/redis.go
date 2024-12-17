package redis

import (
	"QA-System/internal/global/config"
	WeJHSDK "github.com/zjutjh/WeJH-SDK"
)

func getConfig() WeJHSDK.RedisInfoConfig {
	info := WeJHSDK.RedisInfoConfig{
		Host:     "localhost",
		Port:     "6379",
		DB:       0,
		Password: "",
	}
	if global.Config.IsSet("redis.host") {
		info.Host = global.Config.GetString("redis.host")
	}
	if global.Config.IsSet("redis.port") {
		info.Port = global.Config.GetString("redis.port")
	}
	if global.Config.IsSet("redis.db") {
		info.DB = global.Config.GetInt("redis.db")
	}
	if global.Config.IsSet("redis.pass") {
		info.Password = global.Config.GetString("redis.pass")
	}
	return info
}
