package configService

import (
	"QA-system/config/redis"
	"context"
)

var ctx = context.Background()

func GetConfig(key string) string {
	val, err := redis.RedisClient.Get(ctx, key).Result()
	if err == nil {
		return val
	}
	return ""
}
