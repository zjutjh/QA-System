package redis

import (
	"github.com/go-redis/redis/v8"
	WeJHSDK "github.com/zjutjh/WeJH-SDK"
)

var RedisClient *redis.Client
var RedisInfo WeJHSDK.RedisInfoConfig

func init() {
	info := getConfig()

	RedisClient = WeJHSDK.GetRedisClient(info)
	RedisInfo = info

}
