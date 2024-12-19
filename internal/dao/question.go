package dao

import (
	"QA-System/internal/models"
	"QA-System/internal/pkg/redis"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type Question struct {
	ID            int      `json:"id"`
	SerialNum     int      `json:"serial_num"`                                         // 题目序号
	Subject       string   `json:"subject"`                                            // 问题
	Description   string   `json:"description"`                                        // 问题描述
	Img           string   `json:"img"`                                                // 图片
	Required      bool     `json:"required"`                                           // 是否必填
	Unique        bool     `json:"unique"`                                             // 是否唯一
	OtherOption   bool     `json:"other_option"`                                       // 是否有其他选项
	QuestionType  int      `json:"question_type" binding:"required,oneof=1 2 3 4 5 6"` // 问题类型 1单选2多选3填空4简答5图片6文件
	Reg           string   `json:"reg"`                                                // 正则表达式
	Options       []Option `json:"options"`                                            // 选项
	MaximumOption uint     `json:"maximum_option"`                                     // 多选最多选项数 0为不限制
	MinimumOption uint     `json:"minimum_option"`                                     // 多选最少选项数 0为不限制
}

type QuestionsList struct {
	QuestionID int    `json:"question_id" binding:"required"`
	SerialNum  int    `json:"serial_num"`
	Answer     string `json:"answer"`
}

func (d *Dao) CreateQuestion(ctx context.Context, question models.Question) (models.Question, error) {
	err := d.orm.WithContext(ctx).Create(&question).Error
	return question, err
}

func (d *Dao) GetQuestionsBySurveyID(ctx context.Context, surveyID int) ([]models.Question, error) {
	var questions []models.Question
	cacheData, err := redis.RedisClient.Get(ctx, fmt.Sprintf("questions:sid:%d", surveyID)).Result()
	if err == nil && cacheData != "" {
		// 反序列化 JSON 为结构体
		if err := json.Unmarshal([]byte(cacheData), &questions); err == nil {
			return questions, nil
		}
	}
	err = d.orm.WithContext(ctx).Model(models.Question{}).Where("survey_id = ?", surveyID).Find(&questions).Error
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

func (d *Dao) GetQuestionByID(ctx context.Context, questionID int) (*models.Question, error) {
	var question models.Question
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

func (d *Dao) DeleteQuestion(ctx context.Context, questionID int) error {
	err := redis.RedisClient.Del(ctx, fmt.Sprintf("question:qid:%d", questionID)).Err()
	if err != nil {
		return err
	}
	err = d.orm.WithContext(ctx).Where("id = ?", questionID).Delete(&models.Question{}).Error
	return err
}

func (d *Dao) DeleteQuestionBySurveyID(ctx context.Context, surveyID int) error {
	err := redis.RedisClient.Del(ctx, fmt.Sprintf("questions:sid:%d", surveyID)).Err()
	if err != nil {
		return err
	}
	err = d.orm.WithContext(ctx).Where("survey_id = ?", surveyID).Delete(&models.Question{}).Error
	return err
}

func (d *Dao) CreateType(ctx context.Context, name string, value string) error {
	//如果type已经存在则直接更新当前type的value
	var t models.Type
	err := d.orm.WithContext(ctx).Where("type = ?", name).First(&t).Error
	if err == nil {
		err = d.orm.WithContext(ctx).Model(&t).Update("value", value).Error
		return err
	}
	err = d.orm.WithContext(ctx).Create(&models.Type{Type: name, Value: value}).Error
	return err
}

func (d *Dao) GetType(ctx context.Context, name string) (string, error) {
	var t models.Type
	err := d.orm.WithContext(ctx).Where("type = ?", name).First(&t).Error
	return t.Value, err
}

func (d *Dao) DeleteAllQuestionCache(ctx context.Context) error {
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
