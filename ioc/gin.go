package ioc

import (
	"time"

	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/codebook"
	"github.com/Duke1616/ecmdb/internal/dataio"
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
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gotomicro/ego/server/egin"
)

func InitWebServer(sp session.Provider, checkPolicyMiddleware *middleware.CheckPolicyMiddlewareBuilder,
	mdls []gin.HandlerFunc, modelHdl *model.Handler, attributeHdl *attribute.Handler,
	resourceHdl *resource.Handler, rmHdl *relation.RMHandler, rrHdl *relation.RRHandler,
	rtHdl *relation.RTHandler, userHdl *user.Handler, templateHdl *template.Handler, strategyHdl *strategy.Handler,
	codebookHdl *codebook.Handler, runnerHdl *runner.Handler, orderHdl *order.Handler, workflowHdl *workflow.Handler,
	templateGroupHdl *template.GroupHdl, engineHdl *engine.Handler, taskHdl *task.Handler, policyHdl *policy.Handler,
	menuHdl *menu.Handler, endpointHdl *endpoint.Handler, roleHdl *role.Handler, permissionHdl *permission.Handler,
	departmentHdl *department.Handler, toolsHdl *tools.Handler, termHdl *terminal.Handler, rotaHdl *rota.Handler,
	discoveryHdl *discovery.Handler, dataIOHdl *dataio.Handler, checkLoginMiddleware *middleware.CheckLoginMiddlewareBuilder,
) *egin.Component {
	session.SetDefaultProvider(sp)
	gin.SetMode(gin.ReleaseMode)

	server := egin.DefaultContainer().Build(egin.WithPort(8000))
	server.Use(mdls...)

	// 不需要登录认证鉴权的路由
	userHdl.PublicRoutes(server.Engine)
	strategyHdl.PublicRoutes(server.Engine)
	toolsHdl.PublicRoutes(server.Engine)
	orderHdl.PublicRoute(server.Engine)

	// 验证是否登录
	server.Use(session.CheckLoginMiddleware())

	// 查看用户拥有权限
	permissionHdl.PublicRoutes(server.Engine)

	// 检查权限策略
	server.Use(checkPolicyMiddleware.Build())

	// CMDB 相关接口
	modelHdl.PrivateRoutes(server.Engine)
	attributeHdl.PrivateRoutes(server.Engine)
	resourceHdl.PrivateRoutes(server.Engine)
	rmHdl.PrivateRoute(server.Engine)
	rrHdl.PrivateRoute(server.Engine)
	rtHdl.PrivateRoute(server.Engine)
	termHdl.PrivateRoutes(server.Engine)
	dataIOHdl.PrivateRoutes(server.Engine)

	// 工单流程相关接口
	workflowHdl.PrivateRoutes(server.Engine)
	templateGroupHdl.PrivateRoutes(server.Engine)
	discoveryHdl.PrivateRoutes(server.Engine)
	engineHdl.PrivateRoutes(server.Engine)
	orderHdl.PrivateRoutes(server.Engine)
	taskHdl.PrivateRoutes(server.Engine)
	templateHdl.PrivateRoutes(server.Engine)
	codebookHdl.PrivateRoutes(server.Engine)
	runnerHdl.PrivateRoutes(server.Engine)

	// 排班系统相关接口
	rotaHdl.PrivateRoutes(server.Engine)

	// 用户权限相关接口
	userHdl.PrivateRoutes(server.Engine)
	permissionHdl.PrivateRoutes(server.Engine)
	policyHdl.PrivateRoutes(server.Engine)
	menuHdl.PrivateRoutes(server.Engine)
	endpointHdl.PrivateRoutes(server.Engine)
	departmentHdl.PrivateRoutes(server.Engine)
	roleHdl.PrivateRoutes(server.Engine)

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
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{"POST", "GET", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:  []string{"Content-Type", "Authorization", "X-Finder-Id", "X-Finder-ID"},
		ExposeHeaders: []string{"X-Access-Token"},
		// 是否允许你带 cookie 之类的东西
		AllowCredentials: true,
		//AllowOriginFunc: func(origin string) bool {
		//	if strings.HasPrefix(origin, "http://localhost") {
		//		// 你的开发环境
		//		return true
		//	}
		//	return strings.Contains(origin, "example.com")
		//},
		MaxAge: 12 * time.Hour,
	})
}
