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
	"github.com/Duke1616/ecmdb/pkg/contract/permission"
	pluginx "github.com/Duke1616/ecmdb/pkg/plugin/types"
	"github.com/Duke1616/eiam/pkg/ctxutil"
	"github.com/Duke1616/eiam/pkg/web/capability"
	"github.com/ecodeclub/ginx"
	"github.com/gin-gonic/gin"
)

var _ ginx.Handler = &Handler{}

type Handler struct {
	svc pluginservice.Service
	capability.IRegistry
}

func NewHandler(svc pluginservice.Service) *Handler {
	return &Handler{
		svc:       svc,
		IRegistry: capability.NewRegistry("cmdb", "plugin", "插件中心/插件管理"),
	}
}

func (h *Handler) PublicRoutes(server *gin.Engine) {
	server.Any("/api/plugin-runtime/:plugin_id/*any", h.ProxyToPlugin)
}

func (h *Handler) IdentifyRoutes(_ *gin.Engine) {}

func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/plugin")
	g.GET("/list", h.Define("插件目录", "view").
		Needs(permission.Plugin.Enums, permission.Plugin.Get).
		Bind(ginx.W(h.ListPlugins)),
	)
	g.GET("/detail", h.Define("插件详情", "get").
		NoSync().
		Bind(ginx.W(h.GetPluginDetail)),
	)
	g.GET("/enums", h.Define("插件枚举", "enums").
		NoSync().
		Bind(ginx.W(h.ListEnums)),
	)
	g.GET("/definition/default", h.Define("查询默认插件定义", "default").
		NoSync().
		Bind(ginx.W(h.GetDefaultDefinition)),
	)
	g.POST("/bindings/save", h.Define("保存插件绑定", "create").
		Needs(permission.Plugin.Default, permission.Attribute.View, permission.Model.View).
		Bind(ginx.B[SaveBindingsReq](h.SaveBindings)),
	)
	g.PATCH("/binding/switch/:uid", h.Define("切换插件绑定状态", "switch").
		Bind(ginx.W(h.SwitchBindingStatus)),
	)
	g.DELETE("/binding/delete/:uid", h.Define("删除插件绑定", "delete").
		Bind(ginx.W(h.DeleteBinding)),
	)
	g.POST("/resource/actions/batch", h.Define("查询资源插件动作", "actions").
		NoSync().
		Bind(ginx.B[ListResourceActionsBatchReq](h.ListResourceActionsBatch)),
	)
	g.POST("/action/resolve", h.Define("解析插件动作", "resolve").
		NoSync().
		Bind(ginx.B[pluginx.ResolveRequest](h.ResolveAction)),
	)
	g.GET("/runtime/view", h.Define("插件运行时视图", "runtime_view").
		Needs(permission.Plugin.Resolve).
		Bind(ginx.W(h.GetRuntimeView)),
	)
}


func (h *Handler) ListPlugins(ctx *ginx.Context) (ginx.Result, error) {
	items, err := h.svc.ListPlugins(ctx.Context)
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

func (h *Handler) GetPluginDetail(ctx *ginx.Context) (ginx.Result, error) {
	uid := ctx.Query("uid").StringOrDefault("")
	detail, err := h.svc.GetPluginDetail(ctx.Context, uid)
	if err != nil {
		return ginx.Result{Msg: "查询插件详情失败"}, err
	}

	return ginx.Result{
		Msg:  "查询插件详情成功",
		Data: detail,
	}, nil
}

func (h *Handler) ListEnums(ctx *ginx.Context) (ginx.Result, error) {
	items, err := h.svc.ListEnums(ctx.Context)
	if err != nil {
		return ginx.Result{Msg: "查询插件枚举失败"}, err
	}

	return ginx.Result{
		Msg:  "查询插件枚举成功",
		Data: items,
	}, nil
}

func (h *Handler) GetDefaultDefinition(ctx *ginx.Context) (ginx.Result, error) {
	pluginID := ctx.Query("plugin_id").StringOrDefault("")
	def, err := h.svc.GetDefaultDefinition(ctx.Context, pluginID)
	if err != nil {
		return ginx.Result{Msg: "查询默认插件定义失败"}, err
	}

	return ginx.Result{
		Msg:  "查询默认插件定义成功",
		Data: def,
	}, nil
}

func (h *Handler) SaveBindings(ctx *ginx.Context, req SaveBindingsReq) (ginx.Result, error) {
	if err := h.svc.SaveBindings(ctx.Context, domain.SavePluginBindings{
		PluginID: req.PluginID,
		Bindings: req.Bindings,
	}); err != nil {
		return ginx.Result{Msg: "保存插件绑定失败"}, err
	}
	return ginx.Result{Msg: "保存插件绑定成功"}, nil
}

func (h *Handler) SwitchBindingStatus(ctx *ginx.Context) (ginx.Result, error) {
	uid, err := ctx.Param("uid").AsString()
	if err != nil {
		return ginx.Result{Msg: "获取 uid 参数失败"}, err
	}
	enabled, err := h.svc.ToggleBindingStatus(ctx.Context, uid)
	if err != nil {
		return ginx.Result{Msg: "更新插件绑定状态失败"}, err
	}
	return ginx.Result{
		Msg:  "更新插件绑定状态成功",
		Data: map[string]any{"enabled": enabled},
	}, nil
}

func (h *Handler) DeleteBinding(ctx *ginx.Context) (ginx.Result, error) {
	uid, err := ctx.Param("uid").AsString()
	if err != nil {
		return ginx.Result{Msg: "获取 uid 参数失败"}, err
	}
	if err := h.svc.DeleteBinding(ctx.Context, uid); err != nil {
		return ginx.Result{Msg: "删除插件绑定失败"}, err
	}
	return ginx.Result{Msg: "删除插件绑定成功"}, nil
}

func (h *Handler) ListResourceActionsBatch(ctx *ginx.Context, req ListResourceActionsBatchReq) (ginx.Result, error) {
	actions, err := h.svc.ListResourceActionsBatch(ctx.Context, req.ResourceIDs)
	if err != nil {
		return ginx.Result{Msg: "批量查询插件动作失败"}, err
	}

	return ginx.Result{
		Msg:  "批量查询插件动作成功",
		Data: actions,
	}, nil
}

func (h *Handler) ResolveAction(ctx *ginx.Context, req pluginx.ResolveRequest) (ginx.Result, error) {
	result, err := h.svc.ResolveAction(ctx.Context, req)
	if err != nil {
		return ginx.Result{Msg: "解析插件动作失败"}, err
	}

	return ginx.Result{
		Msg:  "解析插件动作成功",
		Data: result,
	}, nil
}

func (h *Handler) GetRuntimeView(ctx *ginx.Context) (ginx.Result, error) {
	resourceID, err := ctx.Query("resource_id").AsInt64()
	if err != nil || resourceID <= 0 {
		return ginx.Result{Msg: "resource_id 非法"}, errs.ValidationError.WithMsg("resource_id 非法")
	}

	pluginID := strings.TrimSpace(ctx.Query("plugin_id").StringOrDefault(""))
	action := strings.TrimSpace(ctx.Query("action").StringOrDefault(""))

	p, spec, err := h.svc.GetActionRuntime(ctx.Context, pluginID, action)
	if err != nil {
		return ginx.Result{Msg: "解析插件运行时失败"}, err
	}

	return ginx.Result{
		Msg:  "解析插件运行时成功",
		Data: buildRuntimeViewFromSpec(p, spec, resourceID),
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

func buildRuntimeViewFromSpec(p domain.Plugin, spec pluginx.ActionSpec, resourceID int64) runtimeView {
	props := map[string]any{
		"resourceId": strconv.FormatInt(resourceID, 10),
	}

	presentation := runtimePresentation{
		Title: spec.Name,
	}

	applyActionRuntime(spec.Runtime, props, &presentation)

	return runtimeView{
		PluginID: p.UID,
		Action:   spec.Action,
		Entry: runtimeEntry{
			Format:        "umd",
			JSURL:         staticAssetURL(p.UID, "index.umd.js", p.Version),
			CSSURL:        staticAssetURL(p.UID, "index.css", p.Version),
			GlobalName:    pluginGlobalName(p.UID),
			ComponentName: "Index",
		},
		Runtime: runtimePayload{
			APIBase: "/api/cmdb/plugin-runtime/" + p.UID,
			Props:   props,
		},
		Presentation: presentation,
	}
}

func staticAssetURL(pluginID string, filename string, version string) string {
	base := "/api/cmdb/plugin-runtime/" + pluginID + "/static/" + filename
	version = strings.TrimSpace(version)
	if version == "" {
		return base
	}
	return base + "?v=" + url.QueryEscape(version)
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

	// 公共路由没有登录态，内置插件固定从系统租户空间读取。
	detail, err := h.svc.GetPluginDetail(ctxutil.WithTenantID(ctx.Request.Context(), ctxutil.SystemTenantID), pluginID)
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

