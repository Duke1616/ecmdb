package ioc

import (
	"fmt"

	"github.com/Duke1616/ecmdb/internal/user/ldapx"
	"github.com/spf13/viper"
)

func InitLdapConfig() ldapx.Config {
	// 定义一个结构体实例
	var cfg ldapx.Config

	// 使用 Unmarshal 函数将配置数据解析到结构体中
	if err := viper.UnmarshalKey("ldap", &cfg); err != nil {
		panic(fmt.Errorf("unable to decode into struct: %v", err))
	}

	return cfg
}
