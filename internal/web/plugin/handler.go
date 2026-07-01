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
	g.GET("/list", h.Capability("插件目录", "view").
		Handle(ginx.Wrap(h.ListPlugins)),
	)
	g.GET("/detail", h.Capability("插件详情", "get").
		Handle(ginx.Wrap(h.GetPluginDetail)),
	)
	g.GET("/enums", h.Capability("插件枚举", "view").
		Handle(ginx.Wrap(h.ListEnums)),
	)
	g.POST("/definition/register", h.Capability("注册插件定义", "definition_register").
		Handle(ginx.WrapBody[pluginx.Definition](h.RegisterDefinition)),
	)
	g.POST("/definition/upsert", h.Capability("保存插件定义", "definition_upsert").
		Handle(ginx.WrapBody[pluginx.Plugin](h.UpsertPlugin)),
	)
	g.POST("/binding/upsert", h.Capability("保存插件绑定", "binding_upsert").
		Handle(ginx.WrapBody[pluginx.Binding](h.UpsertBinding)),
	)
	g.POST("/delete", h.Capability("删除插件", "delete").
		Handle(ginx.WrapBody[DeletePluginReq](h.DeletePlugin)),
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
	g.POST("/sync-default-schema", h.Capability("同步默认模型", "sync_default_schema").
		Handle(ginx.WrapBody[SyncDefaultSchemaReq](h.SyncDefaultSchema)),
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

func (h *Handler) DeletePlugin(ctx *gin.Context, req DeletePluginReq) (ginx.Result, error) {
	if err := h.svc.DeletePlugin(ctx.Request.Context(), req.UID); err != nil {
		return ginx.Result{Msg: "删除插件失败"}, err
	}
	return ginx.Result{Msg: "删除插件成功"}, nil
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

func (h *Handler) SyncDefaultSchema(ctx *gin.Context, req SyncDefaultSchemaReq) (ginx.Result, error) {
	if err := h.svc.SyncDefaultSchema(ctx.Request.Context(), req.PluginID); err != nil {
		return ginx.Result{Msg: "同步默认模型失败"}, err
	}
	return ginx.Result{Msg: "同步默认模型成功"}, nil
}

type SyncDefaultSchemaReq struct {
	PluginID string `json:"plugin_id" binding:"required"`
}
