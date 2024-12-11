package dao

import (
	"QA-System/internal/models"
	"QA-System/internal/pkg/redis"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type Option struct {
	SerialNum   int    `json:"serial_num"`  //选项序号
	Content     string `json:"content"`     //选项内容
	Description string `json:"description"` //选项描述
	Img         string `json:"img"`         //图片
}

func (d *Dao) CreateOption(ctx context.Context, option models.Option) error {
	err := d.orm.WithContext(ctx).Create(&option).Error
	return err
}

func (d *Dao) GetOptionsByQuestionID(ctx context.Context, questionID int) ([]models.Option, error) {
	var options []models.Option
	// 从 Redis 获取
	cachedData, err := redis.RedisClient.Get(ctx, fmt.Sprintf("options:qid:%d", questionID)).Result()
	if err == nil && cachedData != "" {
		// 反序列化 JSON 为结构体
		if err := json.Unmarshal([]byte(cachedData), &options); err == nil {
			return options, nil
		}
	}
	err = d.orm.WithContext(ctx).Model(models.Option{}).Where("question_id = ?", questionID).Find(&options).Error
	if err != nil {
		return nil, err
	}
	// 序列化为 JSON 后存储到 Redis
	jsonData, err := json.Marshal(options)
	if err == nil {
		redis.RedisClient.Set(ctx, fmt.Sprintf("options:qid:%d", questionID), jsonData, 20*time.Minute)
	}
	return options, nil
}

func (d *Dao) DeleteOption(ctx context.Context, questionID int) error {
	err := redis.RedisClient.Del(ctx, fmt.Sprintf("options:qid:%d", questionID)).Err()
	if err != nil {
		return err
	}
	err = d.orm.WithContext(ctx).Where("question_id = ?", questionID).Delete(&models.Option{}).Error
	return err
}

func (d *Dao) GetOptionByQIDAndAnswer(ctx context.Context, qid int, answer string) (*models.Option, error) {
	var option models.Option
	// 从 Redis 获取
	cachedData, err := redis.RedisClient.Get(ctx, fmt.Sprintf("option:qid:%d:answer:%s", qid, answer)).Result()
	if err == nil && cachedData != "" {
		// 反序列化 JSON 为结构体
		if err := json.Unmarshal([]byte(cachedData), option); err == nil {
			return &option, nil
		}
	}
	err = d.orm.WithContext(ctx).Model(models.Option{}).Where("question_id = ?  AND content = ?", qid, answer).First(&option).Error
	if err != nil {
		return nil, err
	}
	// 序列化为 JSON 后存储到 Redis
	jsonData, err := json.Marshal(option)
	if err == nil {
		redis.RedisClient.Set(ctx, fmt.Sprintf("option:qid:%d:answer:%s", qid, answer), jsonData, 20*time.Minute)
	}
	return &option, err
}

func (d *Dao) GetOptionByQIDAndSerialNum(ctx context.Context, qid int, serialNum int) (*models.Option, error) {
	var option models.Option
	// 从 Redis 获取
	cachedData, err := redis.RedisClient.Get(ctx, fmt.Sprintf("option:qid:%d:serial_num:%s", qid, serialNum)).Result()
	if err == nil && cachedData != "" {
		// 反序列化 JSON 为结构体
		if err := json.Unmarshal([]byte(cachedData), option); err == nil {
			return &option, nil
		}
	}
	err = d.orm.WithContext(ctx).Model(models.Option{}).Where("question_id = ?  AND serial_num = ?", qid, serialNum).First(&option).Error
	if err != nil {
		return nil, err
	}
	// 序列化为 JSON 后存储到 Redis
	jsonData, err := json.Marshal(option)
	if err == nil {
		redis.RedisClient.Set(ctx, fmt.Sprintf("option:qid:%d:serial_num:%s", qid, serialNum), jsonData, 20*time.Minute)
	}
	return &option, err
}

func (d *Dao) DeleteAllOptionCache(ctx context.Context) error {
	// 定义 Redis 前缀
	prefix := "option"

	var cursor uint64
	for {
		// 使用 SCAN 扫描匹配的键
		keys, nextCursor, err := redis.RedisClient.Scan(ctx, cursor, fmt.Sprintf("%s*", prefix), 100).Result()
		if err != nil {
			return fmt.Errorf("failed to scan Redis keys with prefix %s: %w", prefix, err)
		}

		// 批量删除匹配的键
		if len(keys) > 0 {
			if err := redis.RedisClient.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("failed to delete Redis keys: %w", err)
			}
		}

		// 如果游标返回为 0，表示扫描完成
		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}

	return nil
}
