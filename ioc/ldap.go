package ioc

import (
	"fmt"
	"github.com/Duke1616/ecmdb/internal/user/ldapx"
	"github.com/spf13/viper"
)

func InitLdapConfig() ldapx.Config {
	// 设置要解析的 YAML 文件路径
	viper.SetConfigFile("config/prod.yaml")

	// 读取并解析配置文件
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	// 定义一个结构体实例
	var cfg ldapx.LdapConfig

	// 使用 Unmarshal 函数将配置数据解析到结构体中
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(fmt.Errorf("unable to decode into struct: %v", err))
	}

	return cfg.Ldap
}
