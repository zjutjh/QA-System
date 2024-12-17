package dao

import (
	"context"
	"time"

	"QA-System/internal/models"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

// Dao 数据访问对象
type Dao struct {
	orm   *gorm.DB
	mongo *mongo.Database
}

// New 实例化数据访问对象
func New(orm *gorm.DB, mongodb *mongo.Database) *Dao {
	return &Dao{
		orm:   orm,
		mongo: mongodb,
	}
}

// Daos 数据访问对象接口
type Daos interface {
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByID(ctx context.Context, id int) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error

	SaveAnswerSheet(ctx context.Context, answerSheet AnswerSheet) error
	GetAnswerSheetBySurveyID(ctx context.Context, surveyID int, pageNum int, pageSize int) (
		[]AnswerSheet, *int64, error)
	DeleteAnswerSheetBySurveyID(ctx context.Context, surveyID int) error

	CreateManage(ctx context.Context, id int, surveyID int) error
	DeleteManage(ctx context.Context, id int, surveyID int) error
	DeleteManageBySurveyID(ctx context.Context, surveyID int) error
	CheckManage(ctx context.Context, id int, surveyID int) error
	GetManageByUIDAndSID(ctx context.Context, uid int, sid int) (*models.Manage, error)
	GetManageByUserID(ctx context.Context, uid int) ([]models.Manage, error)

	CreateOption(ctx context.Context, option *models.Option) error
	GetOptionsByQuestionID(ctx context.Context, questionID int) ([]models.Option, error)
	DeleteOption(ctx context.Context, optionID int) error

	CreateQuestion(ctx context.Context, question *models.Question) error
	GetQuestionsBySurveyID(ctx context.Context, surveyID int) ([]models.Question, error)
	GetQuestionByID(ctx context.Context, questionID int) (*models.Question, error)
	DeleteQuestion(ctx context.Context, questionID int) error
	DeleteQuestionBySurveyID(ctx context.Context, surveyID int) error

	CreateSurvey(ctx context.Context, survey *models.Survey) error
	GetSurveyByID(ctx context.Context, surveyID int) (*models.Survey, error)
	GetSurveyByTitle(ctx context.Context, title string, num, size int) ([]models.Survey, *int64, error)
	DeleteSurvey(ctx context.Context, surveyID int) error
	UpdateSurveyStatus(ctx context.Context, surveyID int, status int) error
	UpdateSurvey(ctx context.Context, id int, title, desc, img string, deadline time.Time) error
	GetAllSurveyByUserID(ctx context.Context, userId int) ([]models.Survey, error)
	IncreaseSurveyNum(ctx context.Context, sid int) error

	SaveRecordSheet(ctx context.Context, answerSheet RecordSheet, sid int) error
	DeleteRecordSheets(ctx context.Context, surveyID int) error
}
