package dao

import (
	"QA-System/internal/models"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type Dao struct {
	orm   *gorm.DB
	mongo *mongo.Database
}

func New(orm *gorm.DB, mongo *mongo.Database) *Dao {
	return &Dao{
		orm:   orm,
		mongo: mongo,
	}
}

type Daos interface {
	// user
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByID(ctx context.Context, id int) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error

	//answer
	SaveAnswerSheet(ctx context.Context, answerSheet AnswerSheet) error
	GetAnswerSheetBySurveyID(ctx context.Context, surveyID int, pageNum int, pageSize int) ([]AnswerSheet, *int64, error)
	DeleteAnswerSheetBySurveyID(ctx context.Context, surveyID int) error

	//manage
	CreateManage(ctx context.Context, id int, surveyID int) error
	DeleteManage(ctx context.Context, id int, surveyID int) error
	DeleteManageBySurveyID(ctx context.Context, surveyID int) error
	CheckManage(ctx context.Context, id int, surveyID int) error
	GetManageByUIDAndSID(ctx context.Context, uid int, sid int) (*models.Manage, error)
	GetManageByUserID(ctx context.Context, uid int) ([]models.Manage, error)

	//option
	CreateOption(ctx context.Context, option *models.Option) error
	GetOptionsByQuestionID(ctx context.Context, questionID int) ([]models.Option, error)
	DeleteOption(ctx context.Context, optionID int) error

	//question
	CreateQuestion(ctx context.Context, question *models.Question) error
	GetQuestionsBySurveyID(ctx context.Context, surveyID int) ([]models.Question, error)
	GetQuestionByID(ctx context.Context, questionID int) (*models.Question, error)
	DeleteQuestion(ctx context.Context, questionID int) error
	DeleteQuestionBySurveyID(ctx context.Context, surveyID int) error

	//survey
	CreateSurvey(ctx context.Context, survey *models.Survey) error
	GetSurveyByID(ctx context.Context, surveyID int) (*models.Survey, error)
	GetSurveyByTitle(ctx context.Context, title string, num, size int) ([]models.Survey, *int64, error)
	DeleteSurvey(ctx context.Context, surveyID int) error
	UpdateSurveyStatus(ctx context.Context, surveyID int, status int) error
	UpdateSurvey(ctx context.Context, id int, title, desc, img string, deadline time.Time) error
	GetAllSurveyByUserID(ctx context.Context, userId int) ([]models.Survey, error)
	IncreaseSurveyNum(ctx context.Context, sid int) error

	//record
	SaveRecordSheet(ctx context.Context, answerSheet RecordSheet, sid int) error
	DeleteRecordSheets(ctx context.Context, surveyID int) error
}
