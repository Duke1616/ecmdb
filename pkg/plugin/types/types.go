package types

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
)

// ── Plugin 与 Action ─────────────────────────────────────────────────────────

type Plugin struct {
	ID      int64          `json:"id" bson:"id"`
	UID     string         `json:"uid" bson:"uid"`
	Name    string         `json:"name" bson:"name"`
	Type    string         `json:"type" bson:"type"`
	Version string         `json:"version" bson:"version"`
	Actions []ActionSpec   `json:"actions" bson:"actions"`
	Meta    map[string]any `json:"meta,omitempty" bson:"meta,omitempty"`
	Ctime   int64          `json:"ctime,omitempty" bson:"ctime,omitempty"`
	Utime   int64          `json:"utime,omitempty" bson:"utime,omitempty"`
}

func (p Plugin) Validate() error {
	if strings.TrimSpace(p.UID) == "" {
		return fmt.Errorf("插件 UID 不能为空")
	}
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("插件名称不能为空")
	}
	return nil
}

func (p Plugin) FindAction(name string) (ActionSpec, bool) {
	return lo.Find(p.Actions, func(action ActionSpec) bool {
		return action.Action == name
	})
}

func (p Plugin) ResourceActions() []ResourceAction {
	return lo.Map(p.Actions, func(action ActionSpec, _ int) ResourceAction {
		return ResourceAction{
			PluginID:   p.UID,
			Action:     action.Action,
			Name:       action.Name,
			Icon:       action.Icon,
			Placement:  action.Placement,
			Permission: action.Permission,
			BindingUID: action.BindingUID,
			Runtime:    action.Runtime,
			Meta:       action.Meta,
		}
	})
}

type RuntimeSpec struct {
	Mode       string `json:"mode"`
	Upstream   string `json:"upstream,omitempty"`
	HealthPath string `json:"health_path,omitempty"`
}

func (p *Plugin) SetRuntime(spec RuntimeSpec) {
	if p.Meta == nil {
		p.Meta = make(map[string]any)
	}
	p.Meta["runtime"] = spec
}

func (p Plugin) Runtime() (RuntimeSpec, bool) {
	value, ok := p.Meta["runtime"]
	if !ok || value == nil {
		return RuntimeSpec{}, false
	}

	switch spec := value.(type) {
	case RuntimeSpec:
		return spec, true
	case *RuntimeSpec:
		if spec == nil {
			return RuntimeSpec{}, false
		}
		return *spec, true
	case map[string]any:
		return runtimeFromMap(spec)
	default:
		return RuntimeSpec{}, false
	}
}

func runtimeFromMap(src map[string]any) (RuntimeSpec, bool) {
	spec := RuntimeSpec{
		Mode:       stringFromMap(src, "mode"),
		Upstream:   stringFromMap(src, "upstream"),
		HealthPath: stringFromMap(src, "health_path"),
	}
	return spec, spec.Mode != "" || spec.Upstream != "" || spec.HealthPath != ""
}

func stringFromMap(src map[string]any, key string) string {
	val, ok := src[key]
	if !ok || val == nil {
		return ""
	}
	return fmt.Sprint(val)
}

type ActionSpec struct {
	Action     string             `json:"action" bson:"action"`
	Name       string             `json:"name" bson:"name"`
	Icon       string             `json:"icon" bson:"icon"`
	Placement  string             `json:"placement" bson:"placement"`
	Permission string             `json:"permission,omitempty" bson:"permission,omitempty"`
	BindingUID string             `json:"binding_uid,omitempty" bson:"binding_uid,omitempty"`
	Runtime    *ActionRuntimeSpec `json:"runtime,omitempty" bson:"runtime,omitempty"`
	Meta       map[string]any     `json:"meta,omitempty" bson:"meta,omitempty"`
}

type ActionRuntimeSpec struct {
	Layout  string              `json:"layout,omitempty" bson:"layout,omitempty"`
	Title   string              `json:"title,omitempty" bson:"title,omitempty"`
	Props   map[string]any      `json:"props,omitempty" bson:"props,omitempty"`
	Sidebar *RuntimeSidebarSpec `json:"sidebar,omitempty" bson:"sidebar,omitempty"`
}

type RuntimeSidebarSpec struct {
	Enabled           *bool                       `json:"enabled,omitempty" bson:"enabled,omitempty"`
	Mode              string                      `json:"mode,omitempty" bson:"mode,omitempty"`
	Title             string                      `json:"title,omitempty" bson:"title,omitempty"`
	SearchPlaceholder string                      `json:"search_placeholder,omitempty" bson:"search_placeholder,omitempty"`
	EmptyText         string                      `json:"empty_text,omitempty" bson:"empty_text,omitempty"`
	Collapsible       *bool                       `json:"collapsible,omitempty" bson:"collapsible,omitempty"`
	Resource          *RuntimeSidebarResourceSpec `json:"resource,omitempty" bson:"resource,omitempty"`
}

type RuntimeSidebarResourceSpec struct {
	ModelUID      string   `json:"model_uid,omitempty" bson:"model_uid,omitempty"`
	TitleField    string   `json:"title_field,omitempty" bson:"title_field,omitempty"`
	SubtitleField string   `json:"subtitle_field,omitempty" bson:"subtitle_field,omitempty"`
	SearchFields  []string `json:"search_fields,omitempty" bson:"search_fields,omitempty"`
	Limit         int      `json:"limit,omitempty" bson:"limit,omitempty"`
}

// ── Binding 与 BindingGraph ───────────────────────────────────────────────────

type Binding struct {
	ID       int64         `json:"id"`
	UID      string        `json:"uid"`
	PluginID string        `json:"plugin_id"`
	ModelUID string        `json:"model_uid"`
	Enabled  bool          `json:"enabled"`
	Graph    *BindingGraph `json:"graph,omitempty"`
	Ctime    int64         `json:"ctime,omitempty"`
	Utime    int64         `json:"utime,omitempty"`
}

func (b Binding) Validate() error {
	if strings.TrimSpace(b.UID) == "" {
		return fmt.Errorf("插件绑定 UID 不能为空")
	}
	if strings.TrimSpace(b.PluginID) == "" {
		return fmt.Errorf("plugin_id 不能为空")
	}
	if strings.TrimSpace(b.ModelUID) == "" {
		return fmt.Errorf("model_uid 不能为空")
	}
	if b.Graph == nil || len(b.Graph.Nodes) == 0 {
		return fmt.Errorf("graph 不能为空")
	}
	return nil
}

type BindingGraph struct {
	EntryNodeID string             `json:"entry_node_id"`
	Nodes       []BindingGraphNode `json:"nodes"`
	Edges       []BindingGraphEdge `json:"edges,omitempty"`
}

type BindingGraphNode struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	ModelUID      string         `json:"model_uid"`
	Cardinality   string         `json:"cardinality"`
	Required      bool           `json:"required"`
	FieldMappings []FieldMapping `json:"field_mappings,omitempty"`
	Filters       []Filter       `json:"filters,omitempty"`
}

type FieldMapping struct {
	Input         string `json:"input"`
	ResourceField string `json:"resource_field"`
	Required      bool   `json:"required,omitempty"`
}

type BindingGraphEdge struct {
	From         string `json:"from"`
	To           string `json:"to"`
	RelationType string `json:"relation_type,omitempty"`
	Direction    string `json:"direction,omitempty"`
}

// ── Schema 与模型定义 ─────────────────────────────────────────────────────────

type Schema struct {
	ModelGroups    []ModelGroupSpec `json:"model_groups,omitempty"`
	Models         []ModelSpec      `json:"models,omitempty"`
	RelationTypes  []RelationType   `json:"relation_types,omitempty"`
	ModelRelations []ModelRelation  `json:"model_relations,omitempty"`
}

type ModelGroupSpec struct {
	Name string `json:"name"`
}

type ModelSpec struct {
	UID             string           `json:"uid"`
	Name            string           `json:"name"`
	Icon            string           `json:"icon,omitempty"`
	GroupName       string           `json:"group_name,omitempty"`
	Builtin         bool             `json:"builtin"`
	AttributeGroups []AttributeGroup `json:"attribute_groups,omitempty"`
}

type AttributeGroup struct {
	Name   string      `json:"name"`
	Index  int64       `json:"index,omitempty"`
	Fields []Attribute `json:"fields,omitempty"`
}

type Attribute struct {
	UID      string `json:"uid"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Option   any    `json:"option,omitempty"`
	Required bool   `json:"required,omitempty"`
	Display  bool   `json:"display,omitempty"`
	Secure   bool   `json:"secure,omitempty"`
	Builtin  bool   `json:"builtin,omitempty"`
	Index    int64  `json:"index,omitempty"`
}

type RelationType struct {
	UID            string `json:"uid"`
	Name           string `json:"name"`
	SourceDescribe string `json:"source_describe,omitempty"`
	TargetDescribe string `json:"target_describe,omitempty"`
}

type ModelRelation struct {
	SourceModelUID  string `json:"source_model_uid"`
	TargetModelUID  string `json:"target_model_uid"`
	RelationTypeUID string `json:"relation_type_uid"`
	Mapping         string `json:"mapping"`
}

// ── ResourceSpec 资源拓扑节点 ─────────────────────────────────────────────────

type ResourceSpec struct {
	Name           string            `json:"name"`                      // 资源名称 (Go struct 中的字段名)
	ModelUID       string            `json:"model_uid"`                 // CMDB 模型 UID
	RelationType   string            `json:"relation_type,omitempty"`   // 关联类型 UID
	Direction      string            `json:"direction,omitempty"`       // 关联查询方向：auto / source / target
	Cardinality    string            `json:"cardinality"`               // 数量基数：one / many
	Required       bool              `json:"required"`                  // 是否必填/必存在
	Fields         map[string]string `json:"fields"`                    // 字段映射 [Go 属性字段名]CMDB 属性字段 UID
	RequiredFields []string          `json:"required_fields,omitempty"` // 必填/非空属性字段列表
	Filters        []Filter          `json:"filters,omitempty"`         // 过滤条件
	Children       []ResourceSpec    `json:"children,omitempty"`        // 子级关联资源列表
}

type Filter struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    any    `json:"value"`
}

// ── Action 解析结果与上下文 ───────────────────────────────────────────────────

type ResourceAction struct {
	PluginID   string             `json:"plugin_id"`
	Action     string             `json:"action"`
	Name       string             `json:"name"`
	Icon       string             `json:"icon"`
	Placement  string             `json:"placement"`
	Permission string             `json:"permission,omitempty"`
	BindingUID string             `json:"binding_uid,omitempty"`
	Runtime    *ActionRuntimeSpec `json:"runtime,omitempty"`
	Meta       map[string]any     `json:"meta,omitempty"`
}

type ResourceActions struct {
	ResourceID int64            `json:"resource_id"`
	Actions    []ResourceAction `json:"actions"`
}

type ResolvedResource struct {
	ResourceID int64                    `json:"resource_id,omitempty"`
	ModelUID   string                   `json:"model_uid,omitempty"`
	Fields     map[string]any           `json:"fields"`
	Children   map[string]ResolvedInput `json:"children,omitempty"`
}

type ResolvedInput struct {
	Name        string             `json:"name"`
	Cardinality string             `json:"cardinality"`
	Resources   []ResolvedResource `json:"resources"`
}

type ResolveRequest struct {
	PluginID   string         `json:"plugin_id"`
	Action     string         `json:"action"`
	ResourceID int64          `json:"resource_id"`
	Params     map[string]any `json:"params,omitempty"`
}

func (req ResolveRequest) Validate() error {
	if strings.TrimSpace(req.PluginID) == "" {
		return fmt.Errorf("plugin_id 不能为空")
	}
	if strings.TrimSpace(req.Action) == "" {
		return fmt.Errorf("action 不能为空")
	}
	if req.ResourceID <= 0 {
		return fmt.Errorf("resource_id 参数错误")
	}
	return nil
}

type ResolveResult struct {
	PluginID      string                   `json:"plugin_id"`
	PluginName    string                   `json:"plugin_name"`
	PluginVersion string                   `json:"plugin_version,omitempty"`
	ActionName    string                   `json:"action_name"`
	Action        string                   `json:"action"`
	Permission    string                   `json:"permission,omitempty"`
	BindingUID    string                   `json:"binding_uid,omitempty"`
	ModelUID      string                   `json:"model_uid,omitempty"`
	ResourceID    int64                    `json:"resource_id"`
	Inputs        map[string]ResolvedInput `json:"inputs"`
	Params        map[string]any           `json:"params,omitempty"`
	Runtime       *ActionRuntimeSpec       `json:"runtime,omitempty"`
	Meta          map[string]any           `json:"meta,omitempty"`
}

type ActionContext struct {
	Plugin     Plugin                   `json:"plugin"`
	Binding    Binding                  `json:"binding"`
	Action     ActionSpec               `json:"action"`
	ResourceID int64                    `json:"resource_id"`
	Inputs     map[string]ResolvedInput `json:"inputs"`
	Params     map[string]any           `json:"params,omitempty"`
}
