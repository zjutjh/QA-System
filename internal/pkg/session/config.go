package session

import (
	"strings"

	"QA-System/internal/global/config"
	WeJHSDK "github.com/zjutjh/WeJH-SDK"
)

type driver string

const (
	Memory driver = "memory"
	Redis  driver = "redis"
)

var defaultName = "wejh-session"

func getConfig() WeJHSDK.SessionInfoConfig {

	wc := WeJHSDK.SessionInfoConfig{}
	wc.Driver = string(Memory)
	if global.Config.IsSet("session.driver") {
		wc.Driver = strings.ToLower(global.Config.GetString("session.driver"))
	}

	wc.Name = defaultName
	if global.Config.IsSet("session.name") {
		wc.Name = strings.ToLower(global.Config.GetString("session.name"))
	}

	wc.SecretKey = "secret"

	wc.RedisConfig = getRedisConfig()

	return wc
}

func getRedisConfig() WeJHSDK.RedisInfoConfig {
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
