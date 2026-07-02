package plugincmd

import (
	"fmt"

	pluginioc "github.com/Duke1616/ecmdb/cmd/plugin/ioc"
	"github.com/Duke1616/eiam/pkg/ctxutil"
	"github.com/spf13/cobra"
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

		// 将 Context 注入为系统级别根租户，使内置插件作为全局共享配置同步
		ctx := ctxutil.WithTenantID(cmd.Context(), ctxutil.SystemTenantID)
		ctx = ctxutil.WithOriginTenantID(ctx, ctxutil.SystemTenantID)

		if err = app.PluginSvc.RegisterBuiltinPlugins(ctx); err != nil {
			return fmt.Errorf("注册内置插件失败: %w", err)
		}
		fmt.Println("内置插件元数据注册完成")
		return nil
	},
}

func init() {
	Cmd.AddCommand(importBuiltinCmd)
}
