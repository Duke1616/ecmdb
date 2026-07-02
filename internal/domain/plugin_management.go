package domain

import pluginx "github.com/Duke1616/ecmdb/pkg/plugin"

type PluginListItem struct {
	ID           int64                `json:"id"`
	UID          string               `json:"uid"`
	Name         string               `json:"name"`
	Type         string               `json:"type"`
	Version      string               `json:"version"`
	ActionCount  int                  `json:"action_count"`
	BindingCount int                  `json:"binding_count"`
	BoundModels  []PluginBoundModel   `json:"bound_models"`
	Actions      []pluginx.ActionSpec `json:"actions"`
	UpdatedAt    int64                `json:"updated_at"`
}

type PluginBoundModel struct {
	UID       string `json:"uid"`
	Name      string `json:"name"`
	GroupName string `json:"group_name,omitempty"`
	Icon      string `json:"icon,omitempty"`
	Builtin   bool   `json:"builtin"`
}

type PluginBindingDetail struct {
	ID        int64                 `json:"id"`
	UID       string                `json:"uid"`
	PluginID  string                `json:"plugin_id"`
	ModelUID  string                `json:"model_uid"`
	ModelName string                `json:"model_name,omitempty"`
	GroupName string                `json:"group_name,omitempty"`
	ModelIcon string                `json:"model_icon,omitempty"`
	Enabled   bool                  `json:"enabled"`
	Graph     *pluginx.BindingGraph `json:"graph,omitempty"`
}

type PluginDetail struct {
	Plugin   pluginx.Plugin        `json:"plugin"`
	Bindings []PluginBindingDetail `json:"bindings"`
}

type PluginManagementEnums struct {
	Types         []string        `json:"types"`
	Placements    []EnumOption    `json:"placements"`
	UIs           []EnumOption    `json:"uis"`
	Directions    []EnumOption    `json:"directions"`
	RelationTypes []EnumOption    `json:"relation_types"`
	Cardinalities []EnumOption    `json:"cardinalities"`
	Mappings      []EnumOption    `json:"mappings"`
	Models        []PluginModelVM `json:"models"`
}

type PluginModelVM struct {
	UID       string `json:"uid"`
	Name      string `json:"name"`
	GroupName string `json:"group_name,omitempty"`
	Icon      string `json:"icon,omitempty"`
	Builtin   bool   `json:"builtin"`
}

type EnumOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type SavePluginBindings struct {
	PluginID string
	Bindings []pluginx.Binding
}
