package queue

import (
	"encoding/json"
	"time"

	"QA-System/internal/dao"
	"github.com/hibiken/asynq"
)

type SubmitSurveyPayload struct {
	ID            int                 `json:"id"`
	Time          string              `json:"time"`
	QuestionsList []dao.QuestionsList `json:"questions_list"`
}

const TypeSubmitSurvey = "survey:submit"

func NewSubmitSurveyTask(id int, questionsList []dao.QuestionsList) (*asynq.Task, error) {
	payload, err := json.Marshal(SubmitSurveyPayload{ID: id, QuestionsList: questionsList, Time: time.Now().Format("2006-01-02 15:04:05")})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSubmitSurvey, payload), nil
}
