package service

import (
	"context"
	"time"

	"QA-System/internal/dao"
	global "QA-System/internal/global/config"
	r "QA-System/internal/pkg/redis"

	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

var (
	ctx = context.Background()
	d   *dao.Dao
)

// Init 函数用于初始化服务。
func Init(db *gorm.DB, mdb *mongo.Database) {
	d = dao.New(db, mdb)
	d.InitializeCache()
}

// GetConfigUrl 获取配置url
func GetConfigUrl() string {
	url := GetRedis("url")
	if url == "" {
		url = global.Config.GetString("url.host")
		SetRedis("url", url)
	}
	return url
}

// GetConfigKey 获取配置key
func GetConfigKey() string {
	key := GetRedis("key")
	if key == "" {
		key = global.Config.GetString("key")
		SetRedis("key", key)
	}
	return key
}

// SetRedis 设置存储在redis的值
func SetRedis(key string, value string) bool {
	t := int64(900)
	expire := time.Duration(t) * time.Second
	if err := r.RedisClient.Set(ctx, key, value, expire).Err(); err != nil {
		return false
	}
	return true
}

// GetRedis 获取存储在redis的值
func GetRedis(key string) string {
	result, err := r.RedisClient.Get(ctx, key).Result()
	if err != nil {
		return ""
	}
	return result
}
