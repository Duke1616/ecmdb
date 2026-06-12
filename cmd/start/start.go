package start

import (
	"github.com/Duke1616/ecmdb/ioc"
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/server"
	"github.com/spf13/cobra"
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

		// 创建 ego 应用实例
		egoApp := ego.New(ego.WithDisableBanner(true))

		// 启动后台任务
		app.StartBackgroundTasks(cmd.Context())

		// 启动服务
		if err := egoApp.Serve(
			func() server.Server {
				return app.Web
			}(),
		).Run(); err != nil {
			elog.Panic("startup", elog.FieldErr(err))
		}

		return nil
	},
}
