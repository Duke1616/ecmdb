package initial

import (
	"context"
	"github.com/Duke1616/ecmdb/cmd/initial/full"
	"github.com/Duke1616/ecmdb/cmd/initial/incr"
	"github.com/Duke1616/ecmdb/cmd/initial/ioc"
	"github.com/spf13/cobra"
)

var (
	debug      bool
	TagVersion string
)

var Cmd = &cobra.Command{
	Use:   "init",
	Short: "初始化应用服务",
	Long:  "初始化应用服务，作为环境演示",
	Run: func(cmd *cobra.Command, args []string) {
		// 初始化 Ioc 注册
		app, err := ioc.InitApp()
		cobra.CheckErr(err)

		// 获取系统版本信息
		currentVersion, err := app.VerSvc.GetVersion(context.Background())
		cobra.CheckErr(err)

		// 判断是执行全量 OR 增量数据
		if currentVersion == "" {
			complete(app)
			increment(app, "v1.0.0")
		} else {
			increment(app, currentVersion)
		}

		// 记录当前版本
		err = app.VerSvc.CreateOrUpdateVersion(context.Background(), TagVersion)
		cobra.CheckErr(err)
	},
}

func complete(app *ioc.App) {
	// 初始化Init
	init := full.NewInitial(app)

	// 初始化菜单
	err := init.InitMenu()
	cobra.CheckErr(err)

	// 初始化用户
	err = init.InitUser()
	cobra.CheckErr(err)

	// 初始化角色
	err = init.InitRole()
	cobra.CheckErr(err)
}

func increment(app *ioc.App, currentVersion string) {
	// 注册所有增量版本信息
	incr.RegisterIncr(app)

	// 执行增量数据
	err := incr.RunIncrementalOperations(currentVersion)
	cobra.CheckErr(err)
}

func init() {
	Cmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "show debug info")
}
