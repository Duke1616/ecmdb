package ioc

import (
	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/codebook"
	"github.com/Duke1616/ecmdb/internal/endpoint"
	"github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/menu"
	"github.com/Duke1616/ecmdb/internal/model"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/permission"
	"github.com/Duke1616/ecmdb/internal/pkg/middleware"
	"github.com/Duke1616/ecmdb/internal/policy"
	"github.com/Duke1616/ecmdb/internal/relation"
	"github.com/Duke1616/ecmdb/internal/resource"
	"github.com/Duke1616/ecmdb/internal/role"
	"github.com/Duke1616/ecmdb/internal/runner"
	"github.com/Duke1616/ecmdb/internal/strategy"
	"github.com/Duke1616/ecmdb/internal/task"
	"github.com/Duke1616/ecmdb/internal/template"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/worker"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
)

func InitWebServer(sp session.Provider, checkPolicyMiddleware *middleware.CheckPolicyMiddlewareBuilder,
	mdls []gin.HandlerFunc, modelHdl *model.Handler, attributeHdl *attribute.Handler,
	resourceHdl *resource.Handler, rmHdl *relation.RMHandler, rrHdl *relation.RRHandler, workerHdl *worker.Handler,
	rtHdl *relation.RTHandler, userHdl *user.Handler, templateHdl *template.Handler, strategyHdl *strategy.Handler,
	codebookHdl *codebook.Handler, runnerHdl *runner.Handler, orderHdl *order.Handler, workflowHdl *workflow.Handler,
	templateGroupHdl *template.GroupHdl, engineHdl *engine.Handler, taskHdl *task.Handler, policyHdl *policy.Handler,
	menuHdl *menu.Handler, endpointHdl *endpoint.Handler, roleHdl *role.Handler, permissionHdl *permission.Handler,
) *gin.Engine {
	session.SetDefaultProvider(sp)
	server := gin.Default()
	server.Use(mdls...)

	// 不需要登录认证鉴权的路由
	userHdl.PublicRoutes(server)

	// 验证是否登录
	server.Use(session.CheckLoginMiddleware())
	permissionHdl.PublicRoutes(server)

	userHdl.PrivateRoutes(server)
	modelHdl.RegisterRoutes(server)
	attributeHdl.RegisterRoutes(server)
	resourceHdl.RegisterRoutes(server)
	rmHdl.RegisterRoute(server)
	rrHdl.RegisterRoute(server)
	rtHdl.RegisterRoute(server)
	templateHdl.RegisterRoutes(server)
	codebookHdl.RegisterRoutes(server)
	workerHdl.RegisterRoutes(server)
	runnerHdl.RegisterRoutes(server)
	strategyHdl.RegisterRoutes(server)
	policyHdl.PublicRoutes(server)
	menuHdl.PublicRoutes(server)
	endpointHdl.PublicRoutes(server)
	roleHdl.PublicRoutes(server)

	// 检查权限策略
	server.Use(checkPolicyMiddleware.Build())
	permissionHdl.PrivateRoutes(server)
	workflowHdl.RegisterRoutes(server)
	templateGroupHdl.RegisterRoutes(server)
	engineHdl.RegisterRoutes(server)
	orderHdl.RegisterRoutes(server)
	taskHdl.RegisterRoutes(server)
	return server
}

func InitGinMiddlewares() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		corsHdl(),
		func(ctx *gin.Context) {
		},
	}
}

func corsHdl() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"POST", "GET"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
		// 你不加这个，前端是拿不到的
		ExposeHeaders: []string{"x-jwt-token", "x-refresh-token"},
		// 是否允许你带 cookie 之类的东西
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				// 你的开发环境
				return true
			}
			return strings.Contains(origin, "yourcompany.com")
		},
		MaxAge: 12 * time.Hour,
	})
}
