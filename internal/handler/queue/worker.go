package queue

import (
	"context"
	"encoding/json"
	"errors"

	"QA-System/internal/service"
	"github.com/hibiken/asynq"
)

func HandleSubmitSurveyTask(ctx context.Context, t *asynq.Task) error {
	var p SubmitSurveyPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	// 提交问卷
	err := service.SubmitSurvey(p.ID, p.QuestionsList, p.Time)
	if err != nil {
		return errors.New("提交问卷失败原因: " + err.Error())
	}

	return nil
}
