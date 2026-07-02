package main

import (
	"fmt"
	"os"

	"github.com/Duke1616/ecmdb/cmd/backup"
	"github.com/Duke1616/ecmdb/cmd/initial"
	plugincmd "github.com/Duke1616/ecmdb/cmd/plugin"
	"github.com/Duke1616/ecmdb/cmd/repair"
	"github.com/Duke1616/ecmdb/cmd/server"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/gotomicro/ego/core/elog"
	git "github.com/purpleclay/gitz"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version string
	cfgFile string
)

func main() {
	ver := version
	if version == "" {
		ver = latestTag()
	}

	fmt.Println("  ______    _____   __  __   _____    ____  ")
	fmt.Println(" |  ____|  / ____| |  \\/  | |  __ \\  |  _ \\ ")
	fmt.Println(" | |__    | |      | \\  / | | |  | | | |_) |")
	fmt.Println(" |  __|   | |      | |\\/| | | |  | | |  _ < ")
	fmt.Println(" | |____  | |____  | |  | | | |__| | | |_) |")
	fmt.Println(" |______|  \\_____| |_|  |_| |_____/  |____/ ")

	// 使用颜色来突出显示
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf(" %s: %s\n", cyan("Service Version"), green(ver))

	// 版本初始化
	initial.TagVersion = ver

	rootCmd := &cobra.Command{
		Use:   "ecmdb",
		Short: "CMDB、工单一体化平台",
		Long:  "CMDB、工单一体化平台",
	}

	// 1. 设置全局配置文件参数
	dir, _ := os.Getwd()
	defaultCfg := dir + "/config/config.yaml"
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", defaultCfg, "配置文件路径")

	// 2. 初始化配置中心
	cobra.OnInitialize(initViper)

	// 3. 注册子命令
	rootCmd.AddCommand(server.NewCommand())
	rootCmd.AddCommand(initial.Cmd)
	rootCmd.AddCommand(backup.Cmd)
	rootCmd.AddCommand(repair.Cmd)
	rootCmd.AddCommand(plugincmd.Cmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// initViper 初始化 Viper 并支持动态监听
func initViper() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}

	// 开启文件监控
	viper.WatchConfig()
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Warning: 配置文件读取失败: %v\n", err)
	} else {
		fmt.Printf("Using config file: %s\n", viper.ConfigFileUsed())
		setLogLevel() // 每次读取/重载都设置日志级别
	}

	// 监听配置变更，支持动态切换日志级别
	viper.OnConfigChange(func(in fsnotify.Event) {
		setLogLevel()
	})
}

// setLogLevel 根据配置文件中的 log.debug 动态调整全局日志级别
func setLogLevel() {
	if viper.GetBool("log.debug") {
		elog.DefaultLogger.SetLevel(elog.DebugLevel)
		elog.DefaultLogger.Debug("已根据配置开启 Debug 日志级别")
	} else {
		elog.DefaultLogger.SetLevel(elog.InfoLevel)
	}
}

func latestTag() string {
	gc, err := git.NewClient()
	if err != nil {
		return ""
	}

	tags, _ := gc.Tags(
		git.WithShellGlob("*.*.*"),
		git.WithSortBy(git.CreatorDateDesc, git.VersionDesc),
		git.WithCount(1),
	)
	if len(tags) == 0 {
		return ""
	}

	return tags[0]
}
