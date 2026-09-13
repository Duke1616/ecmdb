package plugin

import (
	"strings"

	"github.com/Duke1616/ecmdb/pkg/plugin/types"
)

// WorkspaceOption 用于微调工作台运行时配置（与 ActionOption 统一）
type WorkspaceOption = ActionOption

// newDefaultWorkspaceRuntime 创建标准工作台 RuntimeSpec 预设
func newDefaultWorkspaceRuntime(title string, sidebarModelUID string) types.ActionRuntimeSpec {
	enabled := true
	collapsible := true
	title = strings.TrimSpace(title)

	return types.ActionRuntimeSpec{
		Layout: "workspace",
		Title:  title,
		Props:  make(map[string]any),
		Sidebar: &types.RuntimeSidebarSpec{
			Enabled:           &enabled,
			Mode:              "resource_list",
			Title:             title,
			SearchPlaceholder: "搜索名称 / 标识",
			EmptyText:         "暂无匹配的资源数据",
			Collapsible:       &collapsible,
			Resource: &types.RuntimeSidebarResourceSpec{
				ModelUID:      strings.TrimSpace(sidebarModelUID),
				TitleField:    "name",
				SubtitleField: "ip",
				SearchFields:  []string{"name", "ip"},
				Limit:         20,
			},
		},
	}
}

func ensureWorkspaceRuntime(a *types.ActionSpec) *types.ActionRuntimeSpec {
	if a.Runtime == nil {
		spec := newDefaultWorkspaceRuntime(a.Name, "")
		a.Runtime = &spec
	}
	return a.Runtime
}

// Workspace 预设构建器：一行代码生成标准工作台与侧边栏配置
// 预设了满足前端 PluginRuntimeWorkspace.vue 的全部默认行为（标题、搜索、空状态提示、分页等）
func Workspace(title string, sidebarModelUID string, opts ...WorkspaceOption) ActionOption {
	return func(a *types.ActionSpec) {
		spec := newDefaultWorkspaceRuntime(title, sidebarModelUID)
		a.Runtime = &spec
		for _, opt := range opts {
			opt(a)
		}
	}
}

// SidebarTitle 自定义侧边栏的顶部标题（默认与 Action 标题一致）
func SidebarTitle(title string) WorkspaceOption {
	return func(a *types.ActionSpec) {
		spec := ensureWorkspaceRuntime(a)
		if spec.Sidebar != nil {
			spec.Sidebar.Title = strings.TrimSpace(title)
		}
	}
}

// CardFields 设置侧边栏卡片展示的主标题字段与副标题字段（如 "name", "ip"）
func CardFields(titleField, subtitleField string) WorkspaceOption {
	return func(a *types.ActionSpec) {
		spec := ensureWorkspaceRuntime(a)
		if spec.Sidebar != nil && spec.Sidebar.Resource != nil {
			spec.Sidebar.Resource.TitleField = strings.TrimSpace(titleField)
			spec.Sidebar.Resource.SubtitleField = strings.TrimSpace(subtitleField)
		}
	}
}

// SearchFields 自定义前端侧边栏搜索时参与匹配的 CMDB 属性字段
func SearchFields(fields ...string) WorkspaceOption {
	return func(a *types.ActionSpec) {
		spec := ensureWorkspaceRuntime(a)
		if spec.Sidebar != nil && spec.Sidebar.Resource != nil {
			spec.Sidebar.Resource.SearchFields = fields
		}
	}
}

// SidebarLimit 设置侧边栏单页拉取数量（默认 20 条）
func SidebarLimit(limit int) WorkspaceOption {
	return func(a *types.ActionSpec) {
		spec := ensureWorkspaceRuntime(a)
		if spec.Sidebar != nil && spec.Sidebar.Resource != nil {
			spec.Sidebar.Resource.Limit = limit
		}
	}
}

// SidebarDisabled 声明该工作台不需要展示左侧资产导航侧边栏（单资源独占工作台）
func SidebarDisabled() WorkspaceOption {
	return func(a *types.ActionSpec) {
		spec := ensureWorkspaceRuntime(a)
		if spec.Sidebar != nil {
			disabled := false
			spec.Sidebar.Enabled = &disabled
		}
	}
}

// Prop 注入单个传递给前端插件 UMD 组件的初始化属性 (componentProps)
func Prop(key string, value any) WorkspaceOption {
	return func(a *types.ActionSpec) {
		spec := ensureWorkspaceRuntime(a)
		if spec.Props == nil {
			spec.Props = make(map[string]any)
		}
		spec.Props[key] = value
	}
}

// Props 批量注入传递给前端插件 UMD 组件的属性 (componentProps)
func Props(props map[string]any) WorkspaceOption {
	return func(a *types.ActionSpec) {
		spec := ensureWorkspaceRuntime(a)
		if spec.Props == nil {
			spec.Props = make(map[string]any, len(props))
		}
		for k, v := range props {
			spec.Props[k] = v
		}
	}
}

