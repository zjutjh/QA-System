package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"QA-System/internal/model"
	"QA-System/internal/pkg/redis"
)

// 类型别名：保持向后兼容，后续逐步迁移为 model.* 直接引用
type (
	BaseConfig      = model.BaseConfig
	QuestionConfig  = model.QuestionConfig
	QuestionList    = model.QuestionList
	QuestionSetting = model.QuestionSetting
	QuestionsList   = model.QuestionSubmit
)

// CreateQuestion 创建问题
func (d *Dao) CreateQuestion(ctx context.Context, question model.Question) (model.Question, error) {
	err := d.orm.WithContext(ctx).Create(&question).Error
	return question, err
}

// GetQuestionsBySurveyID 根据问卷ID获取问题列表
func (d *Dao) GetQuestionsBySurveyID(ctx context.Context, surveyID int64) ([]model.Question, error) {
	var questions []model.Question
	cacheData, err := redis.RedisClient.Get(ctx, fmt.Sprintf("questions:sid:%d", surveyID)).Result()
	if err == nil && cacheData != "" {
		// 反序列化 JSON 为结构体
		if err := json.Unmarshal([]byte(cacheData), &questions); err == nil {
			return questions, nil
		}
	}
	err = d.orm.WithContext(ctx).Model(model.Question{}).
		Where("survey_id = ?", surveyID).
		Order("serial_num ASC").
		Find(&questions).Error
	if err != nil {
		return nil, err
	}
	// 序列化为 JSON 后存储到 Redis
	jsonData, err := json.Marshal(questions)
	if err == nil {
		redis.RedisClient.Set(ctx, fmt.Sprintf("questions:sid:%d", surveyID), jsonData, 20*time.Minute)
	}
	return questions, err
}

// GetQuestionByID 根据问题ID获取问题
func (d *Dao) GetQuestionByID(ctx context.Context, questionID int) (*model.Question, error) {
	var question model.Question
	cachedData, err := redis.RedisClient.Get(ctx, fmt.Sprintf("question:qid:%d", questionID)).Result()
	if err == nil && cachedData != "" {
		// 反序列化 JSON 为结构体
		if err := json.Unmarshal([]byte(cachedData), &question); err == nil {
			return &question, nil
		}
	}
	err = d.orm.WithContext(ctx).Where("id = ?", questionID).First(&question).Error
	if err != nil {
		return nil, err
	}
	// 序列化为 JSON 后存储到 Redis
	jsonData, err := json.Marshal(question)
	if err == nil {
		redis.RedisClient.Set(ctx, fmt.Sprintf("question:qid:%d", questionID), jsonData, 20*time.Minute)
	}
	return &question, err
}

// DeleteQuestion 删除问题
func (d *Dao) DeleteQuestion(ctx context.Context, questionID int) error {
	err := redis.RedisClient.Del(ctx, fmt.Sprintf("question:qid:%d", questionID)).Err()
	if err != nil {
		return err
	}
	err = d.orm.WithContext(ctx).Where("id = ?", questionID).Delete(&model.Question{}).Error
	return err
}

// DeleteQuestionBySurveyID 根据问卷ID删除问题
func (d *Dao) DeleteQuestionBySurveyID(ctx context.Context, surveyID int64) error {
	err := redis.RedisClient.Del(ctx, fmt.Sprintf("questions:sid:%d", surveyID)).Err()
	if err != nil {
		return err
	}
	err = d.orm.WithContext(ctx).Where("survey_id = ?", surveyID).Delete(&model.Question{}).Error
	return err
}

// CreateType 创建预先类型
func (d *Dao) CreateType(ctx context.Context, name string, value string) error {
	// 如果type已经存在则直接更新当前type的value
	var t model.Pre
	err := d.orm.WithContext(ctx).Where("type = ?", name).First(&t).Error
	if err == nil {
		err = d.orm.WithContext(ctx).Model(&t).Update("value", value).Error
		return err
	}
	err = d.orm.WithContext(ctx).Create(&model.Pre{Type: name, Value: value}).Error
	return err
}

// GetType 获取预先类型
func (d *Dao) GetType(ctx context.Context, name string) (string, error) {
	var t model.Pre
	err := d.orm.WithContext(ctx).Where("type = ?", name).First(&t).Error
	return t.Value, err
}

// DeleteAllQuestionCache 删除所有问题缓存
func DeleteAllQuestionCache(ctx context.Context) error {
	// 定义 Redis 前缀
	prefix := "question"

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
