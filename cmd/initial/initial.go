package initial

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	debug bool
)

var Cmd = &cobra.Command{
	Use:   "init",
	Short: "初始化应用服务",
	Long:  "初始化应用服务，作为环境演示",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		fmt.Print(ctx)
	},
}

func init() {
	Cmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "show debug info")
}
