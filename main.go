package main

import (
	"fmt"
	"github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/Duke1616/ecmdb/ioc"
	"github.com/gotomicro/ego/task/ecron"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	initViper()
	app, err := ioc.InitApp()
	if err != nil {
		panic(err)
	}

	initCronjob(app.Jobs)
	engine.RegisterEvents(app.Event)

	// 获取所有注册的路由
	//routes := app.Web.Routes()
	//for _, route := range routes {
	//	_, err = app.Svc.RegisterEndpoint(context.Background(), endpoint.Endpoint{
	//		Method: route.Method,
	//		Path:   route.Path,
	//	})
	//
	//	if err != nil {
	//		panic(err)
	//	}
	//}

	err = app.Web.Run(":8000")
	panic(err)
}

func initCronjob(jobs []*ecron.Component) {
	type Config struct {
		Enabled bool `mapstructure:"enabled"`
	}

	var cfg Config
	if err := viper.UnmarshalKey("cronjob", &cfg); err != nil {
		panic(fmt.Errorf("unable to decode into struct: %v", err))
	}

	if !cfg.Enabled {
		return
	}

	for _, job := range jobs {
		go job.Start()
	}
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
