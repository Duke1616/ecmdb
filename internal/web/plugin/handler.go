package web

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/Duke1616/ecmdb/internal/domain"
	"github.com/Duke1616/ecmdb/internal/errs"
	pluginservice "github.com/Duke1616/ecmdb/internal/service/plugin"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	mongoxplugin "github.com/Duke1616/ecmdb/pkg/mongox/plugin"
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
	g.POST("/resource/actions/batch", ginx.WrapBody[ListResourceActionsBatchReq](h.ListResourceActionsBatch))
	g.POST("/action/resolve", h.Capability("解析插件动作", "resolve").
		Handle(ginx.WrapBody[pluginx.ResolveRequest](h.ResolveAction)),
	)
	g.GET("/runtime/view", ginx.Wrap(h.GetRuntimeView))

}

func (h *Handler) PublicRoutes(server *gin.Engine) {
	server.Any("/api/plugin-runtime/:plugin_id/*any", h.ProxyToPlugin)
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

func (h *Handler) GetRuntimeView(ctx *gin.Context) (ginx.Result, error) {
	resourceID, err := strconv.ParseInt(ctx.Query("resource_id"), 10, 64)
	if err != nil || resourceID <= 0 {
		return ginx.Result{Msg: "resource_id 非法"}, errs.ValidationError.WithMsg("resource_id 非法")
	}

	req := pluginx.ResolveRequest{
		PluginID:   strings.TrimSpace(ctx.Query("plugin_id")),
		Action:     strings.TrimSpace(ctx.Query("action")),
		ResourceID: resourceID,
	}
	result, err := h.svc.ResolveAction(ctx.Request.Context(), req)
	if err != nil {
		return ginx.Result{Msg: "解析插件运行时失败"}, err
	}

	return ginx.Result{
		Msg:  "解析插件运行时成功",
		Data: buildRuntimeView(result),
	}, nil
}

type SaveBindingsReq struct {
	PluginID string            `json:"plugin_id" binding:"required"`
	Bindings []pluginx.Binding `json:"bindings" binding:"required"`
}

type runtimeView struct {
	PluginID     string              `json:"plugin_id"`
	Action       string              `json:"action"`
	Entry        runtimeEntry        `json:"entry"`
	Runtime      runtimePayload      `json:"runtime"`
	Presentation runtimePresentation `json:"presentation"`
}

type runtimeEntry struct {
	Format        string `json:"format"`
	JSURL         string `json:"js_url"`
	CSSURL        string `json:"css_url,omitempty"`
	GlobalName    string `json:"global_name"`
	ComponentName string `json:"component_name"`
}

type runtimePayload struct {
	APIBase string         `json:"api_base"`
	Props   map[string]any `json:"props"`
}

type runtimePresentation struct {
	Layout  string                      `json:"layout,omitempty"`
	Title   string                      `json:"title,omitempty"`
	Sidebar *pluginx.RuntimeSidebarSpec `json:"sidebar,omitempty"`
}

func buildRuntimeView(result pluginx.ResolveResult) runtimeView {
	props := map[string]any{
		"resourceId": strconv.FormatInt(result.ResourceID, 10),
	}

	presentation := runtimePresentation{
		Title: result.PluginName,
	}

	applyActionRuntime(result.Runtime, props, &presentation)

	return runtimeView{
		PluginID: result.PluginID,
		Action:   result.Action,
		Entry: runtimeEntry{
			Format:        "umd",
			JSURL:         "/api/cmdb/plugin-runtime/" + result.PluginID + "/static/index.umd.js",
			CSSURL:        "/api/cmdb/plugin-runtime/" + result.PluginID + "/static/index.css",
			GlobalName:    pluginGlobalName(result.PluginID),
			ComponentName: "Index",
		},
		Runtime: runtimePayload{
			APIBase: "/api/cmdb/plugin-runtime/" + result.PluginID,
			Props:   props,
		},
		Presentation: presentation,
	}
}

func applyActionRuntime(spec *pluginx.ActionRuntimeSpec, props map[string]any, presentation *runtimePresentation) {
	if spec == nil {
		return
	}

	if strings.TrimSpace(spec.Layout) != "" {
		presentation.Layout = strings.TrimSpace(spec.Layout)
	}
	if strings.TrimSpace(spec.Title) != "" {
		presentation.Title = spec.Title
		props["title"] = spec.Title
	}
	if len(spec.Props) > 0 {
		mergeRuntimeProps(props, spec.Props)
	}
	if spec.Sidebar != nil {
		presentation.Sidebar = spec.Sidebar
	}
}

func mergeRuntimeProps(target map[string]any, source map[string]any) {
	for key, value := range source {
		target[key] = value
	}
}

func pluginGlobalName(pluginID string) string {
	parts := strings.FieldsFunc(pluginID, func(r rune) bool {
		return r == '.' || r == '-'
	})

	var builder strings.Builder
	builder.WriteString("EcmdbPlugin")
	for _, part := range parts {
		if part == "" {
			continue
		}
		builder.WriteString(strings.ToUpper(part[:1]))
		if len(part) > 1 {
			builder.WriteString(part[1:])
		}
	}
	return builder.String()
}

// ProxyToPlugin 通用插件反向代理网关，自动读取对应插件的 upstream 动态执行 HTTP/WebSocket 代理
func (h *Handler) ProxyToPlugin(ctx *gin.Context) {
	pluginID := ctx.Param("plugin_id")
	anyPath := ctx.Param("any")

	// NOTE: 公共路由没有登录态，无租户上下文。内置插件是全局共享数据（tenant_id=0），
	// 需要忽略租户过滤才能正确命中，否则会返回 404
	detail, err := h.svc.GetPluginDetail(mongoxplugin.IgnoreTenantContext(ctx.Request.Context()), pluginID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"msg": "未找到对应的插件定义"})
		return
	}

	runtime, ok := detail.Plugin.Runtime()
	if !ok || runtime.Upstream == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "插件 upstream 运行态地址未配置"})
		return
	}

	// 2. 解析真实的 upstream 物理地址
	targetURL, err := url.Parse(runtime.Upstream)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "解析插件 upstream 地址失败"})
		return
	}

	// 3. 创建反向代理实例
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// 自定义 Director 以重写请求路径与 Host
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// 重写为子插件服务自己监听接收的实际相对路径
		req.URL.Path = anyPath
		req.Host = targetURL.Host

		// 透传插件身份与调用头
		req.Header.Set(pluginx.HeaderPluginID, pluginID)
	}

	// 4. 执行反向代理
	proxy.ServeHTTP(ctx.Writer, ctx.Request)
}
