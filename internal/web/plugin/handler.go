package web

import (
	"github.com/Duke1616/ecmdb/internal/domain"
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
	g.GET("/list", h.Capability("插件目录", "view").
		Handle(ginx.Wrap(h.ListPlugins)),
	)
	g.GET("/detail", h.Capability("插件详情", "get").
		Handle(ginx.Wrap(h.GetPluginDetail)),
	)
	g.GET("/enums", h.Capability("插件枚举", "enums").
		NoSync().
		Handle(ginx.Wrap(h.ListEnums)),
	)
	g.GET("/definition/default", h.Capability("查询默认插件定义", "default").
		NoSync().
		Handle(ginx.Wrap(h.GetDefaultDefinition)),
	)
	g.POST("/bindings/save", h.Capability("保存插件绑定", "create").
		Needs("cmdb:plugin:default", "cmdb:attribute:view", "cmdb:model:list").
		Handle(ginx.WrapBody[SaveBindingsReq](h.SaveBindings)),
	)
	g.PATCH("/binding/switch/:uid", h.Capability("切换插件绑定状态", "switch").
		Handle(ginx.Wrap(h.SwitchBindingStatus)),
	)
	g.DELETE("/binding/delete/:uid", h.Capability("删除插件绑定", "delete").
		Handle(ginx.Wrap(h.DeleteBinding)),
	)
	g.POST("/resource/actions/batch", h.Capability("批量查询资源插件动作", "actions").
		NoSync().
		Handle(ginx.WrapBody[ListResourceActionsBatchReq](h.ListResourceActionsBatch)),
	)
	g.POST("/action/resolve", h.Capability("解析插件动作", "resolve").
		Handle(ginx.WrapBody[pluginx.ResolveRequest](h.ResolveAction)),
	)
}

func (h *Handler) ListPlugins(ctx *gin.Context) (ginx.Result, error) {
	items, err := h.svc.ListPlugins(ctx.Request.Context())
	if err != nil {
		return ginx.Result{Msg: "查询插件目录失败"}, err
	}

	return ginx.Result{
		Msg: "查询插件目录成功",
		Data: map[string]any{
			"list":  items,
			"total": len(items),
		},
	}, nil
}

func (h *Handler) GetPluginDetail(ctx *gin.Context) (ginx.Result, error) {
	uid := ctx.Query("uid")
	detail, err := h.svc.GetPluginDetail(ctx.Request.Context(), uid)
	if err != nil {
		return ginx.Result{Msg: "查询插件详情失败"}, err
	}

	return ginx.Result{
		Msg:  "查询插件详情成功",
		Data: detail,
	}, nil
}

func (h *Handler) ListEnums(ctx *gin.Context) (ginx.Result, error) {
	items, err := h.svc.ListEnums(ctx.Request.Context())
	if err != nil {
		return ginx.Result{Msg: "查询插件枚举失败"}, err
	}

	return ginx.Result{
		Msg:  "查询插件枚举成功",
		Data: items,
	}, nil
}

func (h *Handler) GetDefaultDefinition(ctx *gin.Context) (ginx.Result, error) {
	pluginID := ctx.Query("plugin_id")
	def, err := h.svc.GetDefaultDefinition(ctx.Request.Context(), pluginID)
	if err != nil {
		return ginx.Result{Msg: "查询默认插件定义失败"}, err
	}

	return ginx.Result{
		Msg:  "查询默认插件定义成功",
		Data: def,
	}, nil
}

func (h *Handler) SaveBindings(ctx *gin.Context, req SaveBindingsReq) (ginx.Result, error) {
	if err := h.svc.SaveBindings(ctx.Request.Context(), domain.SavePluginBindings{
		PluginID: req.PluginID,
		Bindings: req.Bindings,
	}); err != nil {
		return ginx.Result{Msg: "保存插件绑定失败"}, err
	}
	return ginx.Result{Msg: "保存插件绑定成功"}, nil
}

func (h *Handler) SwitchBindingStatus(ctx *gin.Context) (ginx.Result, error) {
	enabled, err := h.svc.ToggleBindingStatus(ctx.Request.Context(), ctx.Param("uid"))
	if err != nil {
		return ginx.Result{Msg: "更新插件绑定状态失败"}, err
	}
	return ginx.Result{
		Msg:  "更新插件绑定状态成功",
		Data: map[string]any{"enabled": enabled},
	}, nil
}

func (h *Handler) DeleteBinding(ctx *gin.Context) (ginx.Result, error) {
	uid := ctx.Param("uid")
	if err := h.svc.DeleteBinding(ctx.Request.Context(), uid); err != nil {
		return ginx.Result{Msg: "删除插件绑定失败"}, err
	}
	return ginx.Result{Msg: "删除插件绑定成功"}, nil
}

func (h *Handler) ListResourceActionsBatch(ctx *gin.Context, req ListResourceActionsBatchReq) (ginx.Result, error) {
	actions, err := h.svc.ListResourceActionsBatch(ctx.Request.Context(), req.ResourceIDs)
	if err != nil {
		return ginx.Result{Msg: "批量查询插件动作失败"}, err
	}

	return ginx.Result{
		Msg:  "批量查询插件动作成功",
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

type SaveBindingsReq struct {
	PluginID string            `json:"plugin_id" binding:"required"`
	Bindings []pluginx.Binding `json:"bindings" binding:"required"`
}
