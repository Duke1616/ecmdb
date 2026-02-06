package ioc

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/xen0n/go-workwx"
)

func InitWorkWx() *workwx.WorkwxApp {
	type Config struct {
		// CorpSecret 应用的凭证密钥，必填
		CorpSecret string `mapstructure:"corp_secret"`
		// AgentID 应用 ID，必填
		AgentID int64 `mapstructure:"agent_id"`
		// 企业微信 ID
		CorpID string `mapstructure:"corp_id"`
	}

	var cfg Config
	if err := viper.UnmarshalKey("wechat", &cfg); err != nil {
		panic(fmt.Errorf("unable to decode into structure: %v", err))
	}

	workApp := workwx.New(cfg.CorpID).WithApp(cfg.CorpSecret, cfg.AgentID)

	// refresh token
	go workApp.SpawnAccessTokenRefresher()

	return workApp
}
