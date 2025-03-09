package service

import (
	"context"
	"time"

	"QA-System/internal/pkg/redis"
)

// GetUserLimit 获取用户的对该问卷的访问次数
func GetUserLimit(c context.Context, stu_id string, sid string) (int, error) {
	// 从 redis 中获取用户的对该问卷的访问次数
	item := "survey:" + sid + ":stu_id:" + stu_id
	var limit int
	err := redis.RedisClient.Get(c, item).Scan(&limit)
	return limit, err
}

// SetUserLimit 设置用户的对该问卷的访问次数
func SetUserLimit(c context.Context, stuId string, sid string, limit int) error {
	// 设置用户的对该问卷的访问次数
	item := "survey:" + sid + ":stu_id:" + stuId
	// 获取当前时间和第二天零点的时间
	now := time.Now()
	tomorrow := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0, 0, 0, 0,
		now.Location(),
	).Add(24 * time.Hour)
	duration := time.Until(tomorrow) // 计算当前时间到第二天零点的时间间隔
	err := redis.RedisClient.Set(c, item, limit, duration).Err()
	return err
}

// InscUserLimit 更新用户的对该问卷的访问次数+1
func InscUserLimit(c context.Context, stuId string, sid string) error {
	// 更新用户的对该问卷的访问次数
	item := "survey:" + sid + ":stu_id:" + stuId
	err := redis.RedisClient.Incr(c, item).Err()
	return err
}
