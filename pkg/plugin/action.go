package plugin

const (
	PlacementResourceDetailActions = "resource.detail.actions"

	DirectionToSource = "source"
	DirectionToTarget = "target"

	RelationTypeDefault = "default"
	RelationTypeGroup   = "group"
	RelationTypeBelong  = "belong"
	RelationTypeRun     = "run"

	MappingOneToOne   = "one_to_one"
	MappingOneToMany  = "one_to_many"
	MappingManyToMany = "many_to_many"

	CardinalityOne  = "one"
	CardinalityMany = "many"
)

func ValidRelationType(relationType string) bool {
	switch relationType {
	case RelationTypeDefault, RelationTypeGroup, RelationTypeBelong, RelationTypeRun:
		return true
	default:
		return false
	}
}

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

// ResourceSpec 描述插件依赖的 CMDB 资源拓扑树节点定义
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
