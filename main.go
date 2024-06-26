package main

import (
	"github.com/Duke1616/ecmdb/ioc"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	initViper()
	app, err := ioc.InitApp()
	if err != nil {
		panic(err)
	}

	err = app.Web.Run(":8000")
	panic(err)
}

func initViper() {
	file := pflag.String("config",
		"config/prod.yaml", "配置文件路径")
	pflag.Parse()
	// 直接指定文件路径
	viper.SetConfigFile(*file)
	viper.WatchConfig()
	err := viper.ReadInConfig()

	if err != nil {
		panic(err)
	}
}
