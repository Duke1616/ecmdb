package cmd

import (
	"fmt"

	"github.com/Duke1616/ecmdb/cmd/backup"
	"github.com/Duke1616/ecmdb/cmd/endpoint"
	"github.com/Duke1616/ecmdb/cmd/initial"
	"github.com/Duke1616/ecmdb/cmd/repair"
	"github.com/Duke1616/ecmdb/cmd/start"
	"github.com/fsnotify/fsnotify"
	"github.com/gotomicro/ego/core/elog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	confFile string
)

var rootCmd = &cobra.Command{
	Use:   "ecmdb",
	Short: "CMDB、工单一体化平台",
	Long:  "CMDB、工单一体化平台",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func initAll() {
	// 初始化配置文件
	initViper()
}

func initViper() {
	if confFile != "" {
		viper.SetConfigFile(confFile)
	}

	viper.WatchConfig()
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Warning: 配置文件读取失败: %v\n", err)
	} else {
		setLogLevel()
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

func Execute(version string) {
	// 版本初始化
	initial.TagVersion = version

	// 初始化设置
	cobra.OnInitialize(initAll)
	rootCmd.AddCommand(start.Cmd)
	rootCmd.AddCommand(initial.Cmd)
	rootCmd.AddCommand(backup.Cmd)
	rootCmd.AddCommand(endpoint.Cmd)
	rootCmd.AddCommand(repair.Cmd)
	err := rootCmd.Execute()
	cobra.CheckErr(err)
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&confFile, "config-file", "f", "config/prod.yaml", "the service config from file")
}
