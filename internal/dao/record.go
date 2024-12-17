package dao

import (
	"context"
	"time"

	database "QA-System/internal/pkg/database/mongodb"
	"go.mongodb.org/mongo-driver/bson"
)

type RecordSheet struct {
	StudentID string    `json:"student_id" bson:"student_id"` // 学生ID
	Time      time.Time `json:"time" bson:"time"`             // 答卷时间
}

func (d *Dao) SaveRecordSheet(ctx context.Context, answerSheet RecordSheet, sid int) error {
	_, err := d.mongo.Collection(database.Record).InsertOne(ctx, bson.M{"survey_id": sid, "record": answerSheet})
	return err
}

func (d *Dao) DeleteRecordSheets(ctx context.Context, surveyID int) error {
	_, err := d.mongo.Collection(database.Record).DeleteMany(ctx, bson.M{"survey_id": surveyID})
	return err
}
