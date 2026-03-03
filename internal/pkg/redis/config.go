package redis

import (
	"github.com/redis/go-redis/v9"
	"github.com/zjutjh/WeJH-SDK/redisHelper"
)

// RedisClient Redis客户端
var RedisClient *redis.Client

// Init 显式初始化 Redis 客户端。必须在使用 RedisClient 前调用。
func Init() {
	info := getConfig()
	RedisClient = redisHelper.Init(&info)
}
