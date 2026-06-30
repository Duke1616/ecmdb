package web

import (
	"strconv"

	pluginservice "github.com/Duke1616/ecmdb/internal/service/plugin"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	pluginx "github.com/Duke1616/ecmdb/pkg/plugin"
	"github.com/Duke1616/eiam/pkg/web/capability"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc pluginservice.Service
	capability.IRegistry
}

func NewHandler(svc pluginservice.Service) *Handler {
	return &Handler{
		svc:       svc,
		IRegistry: capability.NewRegistry("cmdb", "plugin", "资产仓库/插件能力"),
	}
}

func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/plugin")
	g.POST("/definition/register", h.Capability("注册插件定义", "definition_register").
		Handle(ginx.WrapBody[pluginx.Definition](h.RegisterDefinition)),
	)
	g.POST("/definition/upsert", h.Capability("保存插件定义", "definition_upsert").
		Handle(ginx.WrapBody[pluginx.Plugin](h.UpsertPlugin)),
	)
	g.POST("/binding/upsert", h.Capability("保存插件绑定", "binding_upsert").
		Handle(ginx.WrapBody[pluginx.Binding](h.UpsertBinding)),
	)
	g.GET("/resource/actions", h.Capability("查询资源插件动作", "resource_actions").
		Handle(ginx.Wrap(h.ListResourceActions)),
	)
	g.POST("/action/resolve", h.Capability("解析插件动作", "action_resolve").
		Handle(ginx.WrapBody[pluginx.ResolveRequest](h.ResolveAction)),
	)
}

func (h *Handler) RegisterDefinition(ctx *gin.Context, req pluginx.Definition) (ginx.Result, error) {
	if err := h.svc.RegisterDefinition(ctx.Request.Context(), req); err != nil {
		return ginx.Result{Msg: "注册插件定义失败"}, err
	}
	return ginx.Result{Msg: "注册插件定义成功"}, nil
}

func (h *Handler) UpsertPlugin(ctx *gin.Context, req pluginx.Plugin) (ginx.Result, error) {
	if err := h.svc.UpsertPlugin(ctx.Request.Context(), req); err != nil {
		return ginx.Result{Msg: "保存插件定义失败"}, err
	}
	return ginx.Result{Msg: "保存插件定义成功"}, nil
}

func (h *Handler) UpsertBinding(ctx *gin.Context, req pluginx.Binding) (ginx.Result, error) {
	if err := h.svc.UpsertBinding(ctx.Request.Context(), req); err != nil {
		return ginx.Result{Msg: "保存插件绑定失败"}, err
	}
	return ginx.Result{Msg: "保存插件绑定成功"}, nil
}

func (h *Handler) ListResourceActions(ctx *gin.Context) (ginx.Result, error) {
	resourceID, err := strconv.ParseInt(ctx.Query("resource_id"), 10, 64)
	if err != nil {
		return ginx.Result{Msg: "resource_id 参数错误"}, err
	}

	actions, err := h.svc.ListResourceActions(ctx.Request.Context(), resourceID)
	if err != nil {
		return ginx.Result{Msg: "查询插件动作失败"}, err
	}

	return ginx.Result{
		Msg:  "查询插件动作成功",
		Data: actions,
	}, nil
}

func (h *Handler) ResolveAction(ctx *gin.Context, req pluginx.ResolveRequest) (ginx.Result, error) {
	result, err := h.svc.ResolveAction(ctx.Request.Context(), req)
	if err != nil {
		return ginx.Result{Msg: "解析插件动作失败"}, err
	}

	return ginx.Result{
		Msg:  "解析插件动作成功",
		Data: result,
	}, nil
}
