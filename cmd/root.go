package cmd

import (
	"github.com/Duke1616/ecmdb/cmd/endpoint"
	"github.com/Duke1616/ecmdb/cmd/initial"
	"github.com/Duke1616/ecmdb/cmd/repair"
	"github.com/Duke1616/ecmdb/cmd/start"
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
	// 直接指定文件路径
	viper.SetConfigFile(confFile)
	viper.WatchConfig()
	err := viper.ReadInConfig()

	if err != nil {
		panic(err)
	}
}

func Execute(version string) {
	// 版本初始化
	initial.TagVersion = version

	// 初始化设置
	cobra.OnInitialize(initAll)
	rootCmd.AddCommand(start.Cmd)
	rootCmd.AddCommand(initial.Cmd)
	rootCmd.AddCommand(endpoint.Cmd)
	rootCmd.AddCommand(repair.Cmd)
	err := rootCmd.Execute()
	cobra.CheckErr(err)
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&confFile, "config-file", "f", "config/prod.yaml", "the service config from file")
}
