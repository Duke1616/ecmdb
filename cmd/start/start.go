package start

import (
	"context"
	"fmt"
	"github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/Duke1616/ecmdb/internal/endpoint"
	"github.com/Duke1616/ecmdb/ioc"
	"github.com/gin-gonic/gin"
	"github.com/gotomicro/ego/task/ecron"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Cmd = &cobra.Command{
	Use:   "start",
	Short: "ecmdb API服务",
	Long:  "启动服务，对外暴露接口",
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := ioc.InitApp()
		if err != nil {
			panic(err)
		}

		initCronjob(app.Jobs)
		engine.RegisterEvents(app.Event)

		err = app.Web.Run(":8000")
		panic(err)
	},
}

// 注册定时任务
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

// 生成端点路由信息、方便菜单权限绑定路由
func initEndpoint(web *gin.Engine, svc endpoint.Service) {
	routes := web.Routes()
	for _, route := range routes {
		_, err := svc.RegisterEndpoint(context.Background(), endpoint.Endpoint{
			Method: route.Method,
			Path:   route.Path,
		})

		if err != nil {
			panic(err)
		}
	}
}
