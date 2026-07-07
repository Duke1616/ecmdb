package server

import (
	"github.com/Duke1616/ecmdb/ioc"
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/server"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "server",
		Short: "启动服务节点",
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
			if err = egoApp.Serve(
				func() server.Server {
					return app.Web
				}(),
				func() server.Server {
					return app.GrpcServer
				}(),
			).Run(); err != nil {
				elog.Panic("startup", elog.FieldErr(err))
			}

			return nil
		},
	}
}
