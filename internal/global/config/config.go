package global

import (
	"log"

	"github.com/spf13/viper"
)

var Config = viper.New()

func init() {
	Config.AddConfigPath("conf")
	Config.SetConfigName("config")
	Config.SetConfigType("yaml")
	Config.AddConfigPath(".")
	Config.WatchConfig() // 自动将配置读入Config变量
	err := Config.ReadInConfig()
	if err != nil {
		log.Fatal("Config not find", err)
	}
}
