package start

import (
	"fmt"
	"github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/Duke1616/ecmdb/ioc"
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

		go func() {
			if err = app.Web.Run(":8000"); err != nil {
				panic(err)
			}
		}()

		err = app.Grpc.Serve()
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
