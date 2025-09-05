package ioc

import (
	"strings"
	"time"

	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/codebook"
	"github.com/Duke1616/ecmdb/internal/department"
	"github.com/Duke1616/ecmdb/internal/discovery"
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
	"github.com/Duke1616/ecmdb/internal/rota"
	"github.com/Duke1616/ecmdb/internal/runner"
	"github.com/Duke1616/ecmdb/internal/strategy"
	"github.com/Duke1616/ecmdb/internal/task"
	"github.com/Duke1616/ecmdb/internal/template"
	"github.com/Duke1616/ecmdb/internal/terminal"
	"github.com/Duke1616/ecmdb/internal/tools"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/worker"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitWebServer(sp session.Provider, checkPolicyMiddleware *middleware.CheckPolicyMiddlewareBuilder,
	mdls []gin.HandlerFunc, modelHdl *model.Handler, attributeHdl *attribute.Handler,
	resourceHdl *resource.Handler, rmHdl *relation.RMHandler, rrHdl *relation.RRHandler, workerHdl *worker.Handler,
	rtHdl *relation.RTHandler, userHdl *user.Handler, templateHdl *template.Handler, strategyHdl *strategy.Handler,
	codebookHdl *codebook.Handler, runnerHdl *runner.Handler, orderHdl *order.Handler, workflowHdl *workflow.Handler,
	templateGroupHdl *template.GroupHdl, engineHdl *engine.Handler, taskHdl *task.Handler, policyHdl *policy.Handler,
	menuHdl *menu.Handler, endpointHdl *endpoint.Handler, roleHdl *role.Handler, permissionHdl *permission.Handler,
	departmentHdl *department.Handler, toolsHdl *tools.Handler, termHdl *terminal.Handler, rotaHdl *rota.Handler,
	discoveryHdl *discovery.Handler, checkLoginMiddleware *middleware.CheckLoginMiddlewareBuilder,
) *gin.Engine {
	session.SetDefaultProvider(sp)
	gin.SetMode(gin.ReleaseMode)

	server := gin.Default()
	server.Use(mdls...)

	// 不需要登录认证鉴权的路由
	userHdl.PublicRoutes(server)
	strategyHdl.PublicRoutes(server)
	toolsHdl.PublicRoutes(server)
	orderHdl.PublicRoute(server)

	termHdl.PrivateRoutes(server)

	// 验证是否登录
	server.Use(session.CheckLoginMiddleware())

	// 查看用户拥有权限
	permissionHdl.PublicRoutes(server)

	// 检查权限策略
	server.Use(checkPolicyMiddleware.Build())

	// CMDB 相关接口
	modelHdl.PrivateRoutes(server)
	attributeHdl.PrivateRoutes(server)
	resourceHdl.PrivateRoutes(server)
	rmHdl.PrivateRoute(server)
	rrHdl.PrivateRoute(server)
	rtHdl.PrivateRoute(server)

	// 工单流程相关接口
	workflowHdl.PrivateRoutes(server)
	templateGroupHdl.PrivateRoutes(server)
	discoveryHdl.PrivateRoutes(server)
	engineHdl.PrivateRoutes(server)
	orderHdl.PrivateRoutes(server)
	taskHdl.PrivateRoutes(server)
	templateHdl.PrivateRoutes(server)
	codebookHdl.PrivateRoutes(server)
	workerHdl.PrivateRoutes(server)
	runnerHdl.PrivateRoutes(server)

	// 排班系统相关接口
	rotaHdl.PrivateRoutes(server)

	// 用户权限相关接口
	userHdl.PrivateRoutes(server)
	permissionHdl.PrivateRoutes(server)
	policyHdl.PrivateRoutes(server)
	menuHdl.PrivateRoutes(server)
	endpointHdl.PrivateRoutes(server)
	departmentHdl.PrivateRoutes(server)
	roleHdl.PrivateRoutes(server)

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
