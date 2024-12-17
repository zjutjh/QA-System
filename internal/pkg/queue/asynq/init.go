package asynq

import (
	"strconv"

	"QA-System/internal/handler/queue"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

var (
	// Client 客户端
	Client *asynq.Client
	// Srv 服务端
	Srv *asynq.Server
)

// Init 初始化asynq
func Init() {
	cfg := NewConfig()
	port := strconv.Itoa(cfg.port)
	Client = asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.host + ":" + port,
		DB:       (cfg.db) + 1,
		Username: cfg.user,
		Password: cfg.password,
	})

	Srv = asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.host + ":" + port,
			DB:       (cfg.db) + 1,
			Username: cfg.user,
			Password: cfg.password,
		},
		asynq.Config{
			Concurrency:    10,                          // 并发数
			RetryDelayFunc: asynq.DefaultRetryDelayFunc, // 重试延迟
		},
	)
	mux := asynq.NewServeMux()
	mux.HandleFunc(queue.TypeSubmitSurvey, queue.HandleSubmitSurveyTask)

	if err := Srv.Run(mux); err != nil {
		zap.L().Fatal(err.Error())
	}
}
