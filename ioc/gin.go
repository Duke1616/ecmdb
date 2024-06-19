package ioc

import (
	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/codebook"
	"github.com/Duke1616/ecmdb/internal/model"
	"github.com/Duke1616/ecmdb/internal/relation"
	"github.com/Duke1616/ecmdb/internal/resource"
	"github.com/Duke1616/ecmdb/internal/template"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
)

func InitWebServer(sp session.Provider, mdls []gin.HandlerFunc, modelHdl *model.Handler, attributeHdl *attribute.Handler,
	resourceHdl *resource.Handler, rmHdl *relation.RMHandler, rrHdl *relation.RRHandler,
	rtHdl *relation.RTHandler, ldapHdl *user.Handler, templateHdl *template.Handler, codebookHdl *codebook.Handler) *gin.Engine {
	session.SetDefaultProvider(sp)
	server := gin.Default()

	server.Use(mdls...)
	ldapHdl.PublicRegisterRoutes(server)

	modelHdl.RegisterRoutes(server)
	attributeHdl.RegisterRoutes(server)
	resourceHdl.RegisterRoutes(server)
	rmHdl.RegisterRoute(server)
	rrHdl.RegisterRoute(server)
	rtHdl.RegisterRoute(server)
	templateHdl.RegisterRoutes(server)
	codebookHdl.RegisterRoutes(server)

	// 验证是否登录
	server.Use(session.CheckLoginMiddleware())

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
