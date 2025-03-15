package dao

import (
	"context"
	"time"

	database "QA-System/internal/pkg/database/mongodb"
	"go.mongodb.org/mongo-driver/bson"
)

// RecordSheet 记录表模型
type RecordSheet struct {
	StudentID string    `json:"student_id" bson:"student_id"` // 学生ID
	Time      time.Time `json:"time" bson:"time"`             // 答卷时间
}

// SaveRecordSheet 将记录直接保存到 MongoDB 集合中
func (d *Dao) SaveRecordSheet(ctx context.Context, answerSheet RecordSheet, sid string) error {
	_, err := d.mongo.Collection(database.Record).InsertOne(ctx, bson.M{"survey_id": sid, "record": answerSheet})
	return err
}

// DeleteRecordSheets 删除记录表
func (d *Dao) DeleteRecordSheets(ctx context.Context, surveyID string) error {
	_, err := d.mongo.Collection(database.Record).DeleteMany(ctx, bson.M{"survey_id": surveyID})
	return err
}
