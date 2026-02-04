package initial

import (
	"context"
	"fmt"

	"github.com/Duke1616/ecmdb/cmd/initial/full"
	"github.com/Duke1616/ecmdb/cmd/initial/incr/register"
	"github.com/Duke1616/ecmdb/cmd/initial/ioc"
	"github.com/spf13/cobra"
)

var (
	debug         bool
	TagVersion    string
	targetVersion string
	dryRun        bool
	forceExec     bool
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

		if dryRun {
			fmt.Printf("🔍 干运行模式 - 预览操作\n")
			fmt.Printf("==================================================\n")
			fmt.Printf("当前版本: %s\n", currentVersion)
			if targetVersion != "" {
				fmt.Printf("目标版本: %s\n", targetVersion)
			}
			fmt.Printf("==================================================\n")
			return
		}

		// 判断是执行全量 OR 增量数据
		if forceExec {
			if targetVersion == "" {
				cobra.CheckErr(fmt.Errorf("强制执行模式必须指定目标版本 (-v/--version)"))
			}
			fmt.Printf("⚠️ 正在强制执行版本 %s 的逻辑（跳过版本检查，不更新数据库版本）...\n", targetVersion)
			register.RegisterIncr(app)
			err := register.ForceExecuteVersion(targetVersion)
			cobra.CheckErr(err)
			fmt.Printf("✅ 强制执行完成\n")
			return
		}

		if currentVersion == "" {
			complete(app)
			if targetVersion != "" {
				incrementToVersion(app, "v1.0.0", targetVersion)
			} else {
				increment(app, "v1.0.0")
			}
		} else {
			if targetVersion != "" {
				incrementToVersion(app, currentVersion, targetVersion)
			} else {
				increment(app, currentVersion)
			}
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
	register.RegisterIncr(app)

	// 执行增量数据
	err := register.RunIncrementalOperations(currentVersion)
	cobra.CheckErr(err)

	fmt.Printf("==================================================\n")
	fmt.Printf("✅ 增量更新完成! 系统已更新到最新版本\n")
}

// incrementToVersion 执行到指定版本的增量更新
func incrementToVersion(app *ioc.App, currentVersion, targetVersion string) {
	fmt.Printf("🔄 开始增量更新到指定版本...\n")
	fmt.Printf("📊 当前版本: %s\n", currentVersion)
	fmt.Printf("🎯 目标版本: %s\n", targetVersion)
	fmt.Printf("==================================================\n")

	// 注册所有增量版本信息
	register.RegisterIncr(app)

	// 执行增量数据到指定版本
	err := register.RunIncrementalOperationsToVersion(currentVersion, targetVersion)
	cobra.CheckErr(err)

	fmt.Printf("==================================================\n")
	fmt.Printf("✅ 增量更新完成! 系统已更新到版本 %s\n", targetVersion)
}

// rollbackCmd 回滚命令
var rollbackCmd = &cobra.Command{
	Use:   "rollback [version]",
	Short: "回滚到指定版本",
	Long:  "回滚系统到指定的版本",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		rollbackVersion := args[0]

		// 初始化 Ioc 注册
		app, err := ioc.InitApp()
		cobra.CheckErr(err)

		// 获取当前版本
		currentVersion, err := app.VerSvc.GetVersion(context.Background())
		cobra.CheckErr(err)

		fmt.Printf("🔄 开始回滚操作...\n")
		fmt.Printf("📊 当前版本: %s\n", currentVersion)
		fmt.Printf("🎯 回滚到版本: %s\n", rollbackVersion)
		fmt.Printf("==================================================\n")

		if dryRun {
			fmt.Printf("🔍 干运行模式 - 预览回滚操作\n")
			fmt.Printf("将回滚以下版本:\n")
			// 这里可以添加预览逻辑
			return
		}

		// 注册所有增量版本信息
		register.RegisterIncr(app)

		// 执行回滚操作
		err = register.RollbackToVersion(currentVersion, rollbackVersion)
		cobra.CheckErr(err)

		fmt.Printf("==================================================\n")
		fmt.Printf("✅ 回滚完成! 系统已回滚到版本 %s\n", rollbackVersion)
	},
}

// listCmd 列表命令
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有可用版本",
	Long:  "列出系统中所有可用的版本信息",
	Run: func(cmd *cobra.Command, args []string) {
		// 初始化 Ioc 注册
		app, err := ioc.InitApp()
		cobra.CheckErr(err)

		// 获取当前版本
		currentVersion, err := app.VerSvc.GetVersion(context.Background())
		cobra.CheckErr(err)

		fmt.Printf("📋 版本信息列表\n")
		fmt.Printf("==================================================\n")
		fmt.Printf("当前版本: %s\n", currentVersion)
		fmt.Printf("==================================================\n")
		fmt.Printf("可用版本:\n")

		// 注册所有增量版本信息
		register.RegisterIncr(app)

		// 获取所有版本并显示
		versions := register.GetAllVersions()
		for i, version := range versions {
			status := "未执行"
			if currentVersion != "" && version <= currentVersion {
				status = "已执行"
			}
			if version == currentVersion {
				status = "当前版本"
			}
			fmt.Printf("%d. %s (%s)\n", i+1, version, status)
		}
	},
}

func init() {
	Cmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "show debug info")
	Cmd.PersistentFlags().StringVarP(&targetVersion, "version", "v", "", "指定目标版本 (例如: v1.2.3)")
	Cmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "预览模式，不执行实际操作")
	Cmd.PersistentFlags().BoolVar(&forceExec, "force", false, "强制执行指定版本的增量逻辑（不更新数据库版本号）")

	// 添加子命令
	Cmd.AddCommand(rollbackCmd)
	Cmd.AddCommand(listCmd)
}
