package dao

import (
	"context"
	"errors"

	database "QA-System/internal/pkg/database/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// Answer 各问题答卷模型
type Answer struct {
	QuestionID int    `json:"question_id" bson:"questionid"` // 问题ID
	SerialNum  int    `json:"serial_num" bson:"serialnum"`   // 问题序号
	Subject    string `json:"subject" bson:"subject"`        // 问题标题
	Content    string `json:"content" bson:"content"`        // 答案内容
}

// AnswerSheet mongodb答卷表模型
type AnswerSheet struct {
	SurveyID string             `json:"survey_id" bson:"surveyid"` // 问卷ID
	AnswerID primitive.ObjectID `json:"answer_id" bson:"_id"`      // 答卷ID
	Time     string             `json:"time" bson:"time"`          // 答卷时间
	Unique   bool               `json:"unique" bson:"unique"`      // 是否唯一
	Answers  []Answer           `json:"answers" bson:"answers"`    // 答案列表
}

// QuestionAnswers 问题答案模型
type QuestionAnswers struct {
	Title        string   `json:"title"`
	QuestionType int      `json:"question_type"`
	Answers      []string `json:"answers"`
}

// AnswersResonse 答案响应模型
type AnswersResonse struct {
	QuestionAnswers []QuestionAnswers    `json:"question_answers"`
	AnswerIDs       []primitive.ObjectID `json:"answer_ids"`
	Time            []string             `json:"time"`
}

// SaveAnswerSheet 将答卷直接保存到 MongoDB 集合中
func (d *Dao) SaveAnswerSheet(ctx context.Context, answerSheet AnswerSheet, qids []int) error {
	// 构建查询条件
	matchConditions := make([]bson.M, 0) // 初始化为空切片
	for _, answer := range answerSheet.Answers {
		if contains(qids, answer.QuestionID) {
			matchConditions = append(matchConditions, bson.M{
				"answers": bson.M{
					"$elemMatch": bson.M{
						"questionid": answer.QuestionID,
						"content":    answer.Content,
					},
				},
			})
		}
	}

	if len(matchConditions) == 0 {
		// 没有符合条件的记录，直接插入新记录
		_, err := d.mongo.Collection(database.QA).InsertOne(ctx, answerSheet)
		if err != nil {
			return err
		}
		return nil
	}

	filter := bson.M{
		"unique": true,
		"$or":    matchConditions,
	}

	var result AnswerSheet
	err := d.mongo.Collection(database.QA).FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// 没有找到符合条件的记录，直接插入新记录
			_, err := d.mongo.Collection(database.QA).InsertOne(ctx, answerSheet)
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}

	// 更新找到的记录，将unique设为false
	update := bson.M{
		"$set": bson.M{"unique": false},
	}
	_, err = d.mongo.Collection(database.QA).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// 新增一条记录
	newAnswerSheet := AnswerSheet{
		SurveyID: answerSheet.SurveyID,
		Time:     answerSheet.Time,
		Unique:   true,
		Answers:  answerSheet.Answers,
	}

	_, err = d.mongo.Collection(database.QA).InsertOne(ctx, newAnswerSheet)
	if err != nil {
		return err
	}

	return nil
}

func contains(arr []int, item int) bool {
	for _, a := range arr {
		if a == item {
			return true
		}
	}
	return false
}

// GetAnswerSheetBySurveyID 根据问卷UUID分页获取答卷
func (d *Dao) GetAnswerSheetBySurveyID(
	ctx context.Context, surveyID string, pageNum int, pageSize int, text string, unique bool) (
	[]AnswerSheet, *int64, error) {
	answerSheets := make([]AnswerSheet, 0)
	filter := bson.M{"surveyid": surveyID}

	// 如果 text 不为空，添加 text 的查询条件
	if text != "" {
		filter["answers.content"] = bson.M{"$regex": text, "$options": "i"} // i 表示不区分大小写
	}

	// 如果 unique 为 true，添加 unique 的查询条件
	if unique {
		filter["unique"] = true
	}

	// 设置总记录数查询过滤条件和选项
	countFilter := filter
	countOpts := options.Count()

	// 执行总记录数查询
	total, err := d.mongo.Collection(database.QA).CountDocuments(ctx, countFilter, countOpts)
	if err != nil {
		return nil, nil, err
	}

	// 查询分页超过总页数
	if pageSize != 0 && int64(pageNum) > total/int64(pageSize)+1 {
		return nil, nil, errors.New("页数超出范围")
	}

	// 设置分页查询选项
	opts := options.Find()
	if pageNum != 0 && pageSize != 0 {
		opts.SetSkip(int64((pageNum - 1) * pageSize)) // 计算要跳过的文档数
		opts.SetLimit(int64(pageSize))                // 设置返回的文档数
	}

	// 执行分页查询
	cur, err := d.mongo.Collection(database.QA).Find(ctx, filter, opts)
	if err != nil {
		return nil, nil, err
	}
	defer func(cur *mongo.Cursor, ctx context.Context) {
		err := cur.Close(ctx)
		if err != nil {
			zap.L().Error("Failed to close cursor", zap.Error(err))
			return
		}
	}(cur, ctx)

	// 迭代查询结果
	for cur.Next(ctx) {
		var answerSheet AnswerSheet
		if err := cur.Decode(&answerSheet); err != nil {
			return nil, nil, err
		}
		answerSheets = append(answerSheets, answerSheet)
	}
	if err := cur.Err(); err != nil {
		return nil, nil, err
	}
	// 返回分页数据和总记录数
	return answerSheets, &total, nil
}

// DeleteAnswerSheetBySurveyID 根据问卷UUID删除答卷
func (d *Dao) DeleteAnswerSheetBySurveyID(ctx context.Context, surveyID string) error {
	filter := bson.M{"surveyid": surveyID}
	// 删除所有满足条件的文档
	_, err := d.mongo.Collection(database.QA).DeleteMany(ctx, filter)
	return err
}

// DeleteAnswerSheetByAnswerID 根据答卷ID删除答卷
func (d *Dao) DeleteAnswerSheetByAnswerID(ctx context.Context, answerID primitive.ObjectID) error {
	filter := bson.M{"_id": answerID}
	// 删除对应答卷
	_, err := d.mongo.Collection(database.QA).DeleteOne(ctx, filter)
	return err
}

// GetAnswerSheetByAnswerID 根据答卷ID获取答卷
func (d *Dao) GetAnswerSheetByAnswerID(ctx context.Context, answerID primitive.ObjectID) error {
	var answerSheet AnswerSheet
	filter := bson.M{"_id": answerID}
	err := d.mongo.Collection(database.QA).FindOne(ctx, filter).Decode(&answerSheet)
	return err
}
