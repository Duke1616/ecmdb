package ioc

import (
	"fmt"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	"github.com/spf13/viper"
)

func InitLarkClient() *lark.Client {
	type Config struct {
		AppId     string `mapstructure:"app_id"`
		AppSecret string `mapstructure:"app_secret"`
	}

	var cfg Config
	if err := viper.UnmarshalKey("lark", &cfg); err != nil {
		panic(fmt.Errorf("unable to decode into structure: %v", err))
	}

	return lark.NewClient(cfg.AppId, cfg.AppSecret)
}
