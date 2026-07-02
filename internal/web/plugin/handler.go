package web

import (
	"strconv"

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
	g.GET("/enums", h.Capability("插件枚举", "view").
		Handle(ginx.Wrap(h.ListEnums)),
	)
	g.GET("/definition/default", h.Capability("查询默认插件定义", "binding_upsert").
		Handle(ginx.Wrap(h.GetDefaultDefinition)),
	)
	g.POST("/bindings/save", h.Capability("保存插件绑定", "binding_upsert").
		Handle(ginx.WrapBody[SaveBindingsReq](h.SaveBindings)),
	)
	g.POST("/binding/enabled", h.Capability("切换插件绑定状态", "binding_upsert").
		Handle(ginx.WrapBody[UpdateBindingEnabledReq](h.UpdateBindingEnabled)),
	)
	g.GET("/resource/actions", h.Capability("查询资源插件动作", "resource_actions").
		Handle(ginx.Wrap(h.ListResourceActions)),
	)
	g.GET("/model/actions", h.Capability("查询模型插件动作", "model_actions").
		Handle(ginx.Wrap(h.ListModelActions)),
	)
	g.POST("/action/resolve", h.Capability("解析插件动作", "action_resolve").
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

func (h *Handler) UpdateBindingEnabled(ctx *gin.Context, req UpdateBindingEnabledReq) (ginx.Result, error) {
	if err := h.svc.UpdateBindingEnabled(ctx.Request.Context(), req.UID, req.Enabled); err != nil {
		return ginx.Result{Msg: "更新插件绑定状态失败"}, err
	}
	return ginx.Result{Msg: "更新插件绑定状态成功"}, nil
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

func (h *Handler) ListModelActions(ctx *gin.Context) (ginx.Result, error) {
	modelUID := ctx.Query("model_uid")

	actions, err := h.svc.ListModelActions(ctx.Request.Context(), modelUID)
	if err != nil {
		return ginx.Result{Msg: "查询模型插件动作失败"}, err
	}

	return ginx.Result{
		Msg:  "查询模型插件动作成功",
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
