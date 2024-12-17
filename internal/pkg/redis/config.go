package redis

import (
	"github.com/go-redis/redis/v8"
	WeJHSDK "github.com/zjutjh/WeJH-SDK"
)

// RedisClient Redis客户端
var RedisClient *redis.Client

// RedisInfo Redis配置信息
var RedisInfo WeJHSDK.RedisInfoConfig

func init() {
	info := getConfig()

	RedisClient = WeJHSDK.GetRedisClient(info)
	RedisInfo = info
}
