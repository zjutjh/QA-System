package redis

import (
	"QA-System/internal/global/config"

	"github.com/redis/go-redis/v9"
	"github.com/zjutjh/WeJH-SDK/redisHelper"
)

// RedisClient Redis客户端
var (
	RedisClient *redis.Client
	// Redis Stream暂时用不上了
	// StreamName  = config.Config.GetString("redis.stream_name")
	// GroupName   = config.Config.GetString("redis.group_name")
)

// getConfig 获取 Redis 配置
func getConfig() redisHelper.InfoConfig {
	info := redisHelper.InfoConfig{
		Host:     "localhost",
		Port:     "6379",
		DB:       0,
		Password: "",
	}
	if config.Config.IsSet("redis.host") {
		info.Host = config.Config.GetString("redis.host")
	}
	if config.Config.IsSet("redis.port") {
		info.Port = config.Config.GetString("redis.port")
	}
	if config.Config.IsSet("redis.db") {
		info.DB = config.Config.GetInt("redis.db")
	}
	if config.Config.IsSet("redis.pass") {
		info.Password = config.Config.GetString("redis.pass")
	}
	return info
}
