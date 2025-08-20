package endpoint

import (
	"context"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/endpoint"
	"github.com/Duke1616/ecmdb/ioc"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "endpoint",
	Short: "ecmdb endpoint",
	Long:  "注册所有路由信息到 Endpoint 中，用于动态菜单API鉴权中使用",
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := ioc.InitApp()
		if err != nil {
			panic(err)
		}

		err = initEndpoint(app.Web, app.Svc)
		fmt.Print(err)
		panic(err)
	},
}

// 生成端点路由信息、方便菜单权限绑定路由
func initEndpoint(web *gin.Engine, svc endpoint.Service) error {
	routes := web.Routes()
	for _, route := range routes {
		_, err := svc.RegisterEndpoint(context.Background(), endpoint.Endpoint{
			Method: route.Method,
			Path:   route.Path,
		})

		if err != nil {
			return err
		}
	}

	return nil
}
