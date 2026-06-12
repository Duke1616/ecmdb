package ioc

import (
	"context"
	"net"
	"time"

	attribute "github.com/Duke1616/ecmdb/internal/web/attribute"
	dataio "github.com/Duke1616/ecmdb/internal/web/dataio"
	model "github.com/Duke1616/ecmdb/internal/web/model"
	relation "github.com/Duke1616/ecmdb/internal/web/relation"
	resource "github.com/Duke1616/ecmdb/internal/web/resource"
	terminal "github.com/Duke1616/ecmdb/internal/web/terminal"
	tools "github.com/Duke1616/ecmdb/internal/web/tools"
	"github.com/Duke1616/eiam/pkg/web/capability"
	"github.com/Duke1616/eiam/pkg/web/middleware"
	"github.com/Duke1616/eiam/pkg/web/sdk"
	"github.com/gin-gonic/gin"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/server/egin"
)

func InitWebServer(mdls []gin.HandlerFunc, sdk *sdk.SDK, syncer capability.Syncer, providers []capability.PermissionProvider,
	modelHdl *model.Handler, attributeHdl *attribute.Handler, resourceHdl *resource.Handler,
	rmHdl *relation.RelationTypeHandler,
	toolsHdl *tools.Handler, termHdl *terminal.Handler,
	dataIOHdl *dataio.Handler, listener net.Listener,
) *egin.Component {

	server := egin.Load("server.egin").Build(egin.WithListener(listener))
	// 开启 ContextWithFallback：使 ctx.Context.Value() 自动 fallback 到 ctx.Request.Context().Value()
	server.Engine.ContextWithFallback = true
	server.Use(mdls...)

	// 不需要登录认证鉴权的路由

	// 登录检查
	server.Use(sdk.CheckLogin())

	// 权限策略检查
	server.Use(sdk.CheckPolicy())

	// CMDB 相关接口
	modelHdl.PrivateRoutes(server.Engine)
	attributeHdl.PrivateRoutes(server.Engine)
	resourceHdl.PrivateRoutes(server.Engine)
	rmHdl.PrivateRoute(server.Engine)
	termHdl.PrivateRoutes(server.Engine)
	toolsHdl.PrivateRoutes(server.Engine)
	dataIOHdl.PrivateRoutes(server.Engine)

	// 异步启动 EIAM 资产注册控制器
	go func() {
		// 延迟执行，确保路由完全就绪
		time.Sleep(time.Second)

		// 新版本 SDK 内部会启动后台协程维持租约，需传入长生命周期的 Context
		if err := syncer.WithOption(
			capability.WithPermissions(providers...),
			capability.WithRouter(server.Engine),
		).Sync(context.Background()); err != nil {
			elog.Error("EIAM 资产注册控制器启动失败", elog.FieldErr(err))
		}
	}()

	return server
}

func InitGinMiddlewares() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.AccessLogger(),
		middleware.NewCorsBuilder().Build(),
	}
}
