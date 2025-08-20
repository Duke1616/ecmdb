package ioc

import (
	"fmt"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	"github.com/spf13/viper"
)

func InitFeishu() *lark.Client {
	type Config struct {
		AppId     string `mapstructure:"appId"`
		AppSecret string `mapstructure:"appSecret"`
	}

	var cfg Config
	if err := viper.UnmarshalKey("feishu", &cfg); err != nil {
		panic(fmt.Errorf("unable to decode into struct: %v", err))
	}

	return lark.NewClient(cfg.AppId, cfg.AppSecret)
}
