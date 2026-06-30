package plugincmd

import (
	"fmt"

	pluginioc "github.com/Duke1616/ecmdb/cmd/plugin/ioc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Cmd = &cobra.Command{
	Use:   "plugin",
	Short: "管理插件定义",
}

var importBuiltinCmd = &cobra.Command{
	Use:   "import-builtin",
	Short: "导入系统内置插件",
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := pluginioc.InitApp()
		if err != nil {
			return fmt.Errorf("初始化插件导入服务失败: %w", err)
		}

		ctx, err := app.TenantProvider.Context(cmd.Context())
		if err != nil {
			return err
		}

		if err = app.PluginSvc.SyncBuiltinDefinitions(ctx); err != nil {
			return fmt.Errorf("导入内置插件失败: %w", err)
		}
		fmt.Println("内置插件导入完成")
		return nil
	},
}

func init() {
	importBuiltinCmd.Flags().String("access-key", "", "eiam 租户 access key")
	importBuiltinCmd.Flags().String("secret-key", "", "eiam 租户 secret key")

	// 使用 BindPFlag 将命令行参数绑定至 viper 对应的配置 Key，从而使得 ioc 无参初始化正常运作
	_ = viper.BindPFlag("plugin.builtin.tenant.access_key", importBuiltinCmd.Flags().Lookup("access-key"))
	_ = viper.BindPFlag("plugin.builtin.tenant.secret_key", importBuiltinCmd.Flags().Lookup("secret-key"))

	Cmd.AddCommand(importBuiltinCmd)
}
