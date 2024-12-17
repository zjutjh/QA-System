package asynq

import global "QA-System/internal/global/config"

type Config struct {
	host     string
	port     int
	db       int
	user     string
	password string
}

func NewConfig() *Config {
	return &Config{
		host:     global.Config.GetString("redis.host"),
		port:     global.Config.GetInt("redis.port"),
		db:       global.Config.GetInt("redis.db"),
		user:     global.Config.GetString("redis.user"),
		password: global.Config.GetString("redis.pass"),
	}
}
