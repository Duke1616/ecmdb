package ioc

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/xen0n/go-workwx"
)

func InitWorkWx(viper *viper.Viper) *workwx.WorkwxApp {
	type Config struct {
		// CorpSecret 应用的凭证密钥，必填
		CorpSecret string `yaml:"corpSecret"`
		// AgentID 应用 ID，必填
		AgentID int64 `json:"agentId"`
		// 企业微信 ID
		CorpID string `yaml:"corpId"`
	}

	var cfg Config
	if err := viper.UnmarshalKey("wechat", &cfg); err != nil {
		panic(fmt.Errorf("unable to decode into struct: %v", err))
	}

	return workwx.New(cfg.CorpID).WithApp(cfg.CorpSecret, cfg.AgentID)
}
