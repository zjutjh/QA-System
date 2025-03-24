package dao

import (
	"context"
	"time"

	database "QA-System/internal/pkg/database/mongodb"
	"go.mongodb.org/mongo-driver/bson"
)

// RecordSheet 记录表模型
type RecordSheet struct {
	College      string    `json:"college" bson:"college"`               // 学院
	Name         string    `json:"name" bson:"name"`                     // 姓名
	StudentID    string    `json:"student_id" bson:"student_id"`         // 学生ID
	UserType     string    `json:"user_type" bson:"user_type"`           // 用户类型id
	UserTypeDesc string    `json:"user_type_desc" bson:"user_type_desc"` // 用户类型 text
	Gender       string    `json:"gender" bson:"gender"`                 // 性别
	Time         time.Time `json:"time" bson:"time"`                     // 答卷时间
}

// SaveRecordSheet 将记录直接保存到 MongoDB 集合中
func (d *Dao) SaveRecordSheet(ctx context.Context, answerSheet RecordSheet, sid int) error {
	_, err := d.mongo.Collection(database.Record).InsertOne(ctx, bson.M{"survey_id": sid, "record": answerSheet})
	return err
}

// DeleteRecordSheets 删除记录表
func (d *Dao) DeleteRecordSheets(ctx context.Context, surveyID int) error {
	_, err := d.mongo.Collection(database.Record).DeleteMany(ctx, bson.M{"survey_id": surveyID})
	return err
}
