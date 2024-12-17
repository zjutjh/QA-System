package redis

import (
	"QA-System/internal/global/config"
	WeJHSDK "github.com/zjutjh/WeJH-SDK"
)

func getConfig() WeJHSDK.RedisInfoConfig {
	Info := WeJHSDK.RedisInfoConfig{
		Host:     "localhost",
		Port:     "6379",
		DB:       0,
		Password: "",
	}
	if global.Config.IsSet("redis.host") {
		Info.Host = global.Config.GetString("redis.host")
	}
	if global.Config.IsSet("redis.port") {
		Info.Port = global.Config.GetString("redis.port")
	}
	if global.Config.IsSet("redis.db") {
		Info.DB = global.Config.GetInt("redis.db")
	}
	if global.Config.IsSet("redis.pass") {
		Info.Password = global.Config.GetString("redis.pass")
	}
	return Info
}
