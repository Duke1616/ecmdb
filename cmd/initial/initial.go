package initial

import (
	"context"
	"fmt"

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
	},
}

func complete(app *ioc.App) {
	fmt.Printf("🚀 开始全量初始化系统数据...\n")
	fmt.Printf("==================================================\n")

	// 初始化Init
	init := full.NewInitial(app)

	// 初始化菜单
	fmt.Printf("📋 步骤 1/4: 初始化菜单数据\n")
	err := init.InitMenu()
	cobra.CheckErr(err)

	// 初始化用户
	fmt.Printf("👤 步骤 2/4: 初始化用户数据\n")
	userId, err := init.InitUser()
	cobra.CheckErr(err)

	// 初始化角色
	fmt.Printf("🔐 步骤 3/4: 初始化角色数据\n")
	err = init.InitRole()
	cobra.CheckErr(err)

	// 初始化权限
	fmt.Printf("🔑 步骤 4/4: 初始化权限数据\n")
	err = init.InitPermission(userId)
	cobra.CheckErr(err)

	fmt.Printf("==================================================\n")
	fmt.Printf("🎉 全量初始化完成! 系统已准备就绪\n")
}

func increment(app *ioc.App, currentVersion string) {
	fmt.Printf("🔄 开始增量更新系统数据...\n")
	fmt.Printf("📊 当前版本: %s\n", currentVersion)
	fmt.Printf("==================================================\n")

	// 注册所有增量版本信息
	incr.RegisterIncr(app)

	// 执行增量数据
	err := incr.RunIncrementalOperations(currentVersion)
	cobra.CheckErr(err)

	fmt.Printf("==================================================\n")
	fmt.Printf("✅ 增量更新完成! 系统已更新到最新版本\n")
}

func init() {
	Cmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "show debug info")
}
