package redis

import (
	"github.com/redis/go-redis/v9"
	"github.com/zjutjh/WeJH-SDK/redisHelper"
)

// RedisClient Redis客户端
var RedisClient *redis.Client

func init() {
	info := getConfig()

	RedisClient = redisHelper.Init(&info)
}
