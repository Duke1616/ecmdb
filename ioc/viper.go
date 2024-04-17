package ioc

import (
	"fmt"
	"github.com/spf13/viper"
)

func InitViper() *viper.Viper {
	v := viper.New()
	v.SetConfigFile("config/prod.yaml")

	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	return v
}
