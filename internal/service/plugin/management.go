package plugin

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/Duke1616/ecmdb/internal/domain"
	"github.com/Duke1616/ecmdb/internal/errs"
	pluginx "github.com/Duke1616/ecmdb/pkg/plugin"
	"github.com/samber/lo"
)

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
	ID        int64                  `json:"id"`
	UID       string                 `json:"uid"`
	PluginID  string                 `json:"plugin_id"`
	ModelUID  string                 `json:"model_uid"`
	ModelName string                 `json:"model_name,omitempty"`
	GroupName string                 `json:"group_name,omitempty"`
	ModelIcon string                 `json:"model_icon,omitempty"`
	Enabled   bool                   `json:"enabled"`
	Specs     []pluginx.ResourceSpec `json:"specs"`
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

func (s *service) ListPlugins(ctx context.Context) ([]PluginListItem, error) {
	plugins, err := s.repo.ListPlugins(ctx)
	if err != nil {
		return nil, err
	}
	if len(plugins) == 0 {
		return []PluginListItem{}, nil
	}

	pluginIDs := lo.Map(plugins, func(item domain.Plugin, _ int) string {
		return item.UID
	})
	bindings, err := s.repo.ListBindingsByPluginIDs(ctx, pluginIDs)
	if err != nil {
		return nil, err
	}

	modelMeta, err := s.loadModelMetaByUID(ctx, lo.Uniq(lo.Map(bindings, func(item domain.PluginBinding, _ int) string {
		return item.ModelUID
	})))
	if err != nil {
		return nil, err
	}

	bindingsByPluginID := lo.GroupBy(bindings, func(item domain.PluginBinding) string {
		return item.PluginID
	})

	items := make([]PluginListItem, 0, len(plugins))
	for _, item := range plugins {
		pluginBindings := bindingsByPluginID[item.UID]
		items = append(items, PluginListItem{
			ID:           item.ID,
			UID:          item.UID,
			Name:         item.Name,
			Type:         item.Type,
			Version:      item.Version,
			ActionCount:  len(item.Actions),
			BindingCount: len(pluginBindings),
			BoundModels:  buildBoundModels(pluginBindings, modelMeta),
			Actions:      item.Actions,
			UpdatedAt:    item.Utime,
		})
	}
	return items, nil
}

func (s *service) GetPluginDetail(ctx context.Context, uid string) (PluginDetail, error) {
	uid = strings.TrimSpace(uid)
	if uid == "" {
		return PluginDetail{}, fmt.Errorf("plugin uid 不能为空")
	}

	plugin, err := s.repo.GetPlugin(ctx, uid)
	if err != nil {
		return PluginDetail{}, err
	}

	bindings, err := s.repo.ListBindingsByPluginID(ctx, uid)
	if err != nil {
		return PluginDetail{}, err
	}

	modelMeta, err := s.loadModelMetaByUID(ctx, lo.Uniq(lo.Map(bindings, func(item domain.PluginBinding, _ int) string {
		return item.ModelUID
	})))
	if err != nil {
		return PluginDetail{}, err
	}

	details := make([]PluginBindingDetail, 0, len(bindings))
	for _, binding := range bindings {
		meta := modelMeta[binding.ModelUID]
		details = append(details, PluginBindingDetail{
			ID:        binding.ID,
			UID:       binding.UID,
			PluginID:  binding.PluginID,
			ModelUID:  binding.ModelUID,
			ModelName: meta.Name,
			GroupName: meta.GroupName,
			ModelIcon: meta.Icon,
			Enabled:   binding.Enabled,
			Specs:     binding.Specs,
		})
	}

	return PluginDetail{
		Plugin:   plugin,
		Bindings: details,
	}, nil
}

func (s *service) ListEnums(ctx context.Context) (PluginManagementEnums, error) {
	models, err := s.models.ListAll(ctx)
	if err != nil {
		return PluginManagementEnums{}, err
	}

	groupMap, err := s.loadModelGroupNameMap(ctx, models)
	if err != nil {
		return PluginManagementEnums{}, err
	}

	modelItems := lo.Map(models, func(item domain.Model, _ int) PluginModelVM {
		return PluginModelVM{
			UID:       item.UID,
			Name:      item.Name,
			GroupName: groupMap[item.GroupId],
			Icon:      item.Icon,
			Builtin:   item.Builtin,
		}
	})
	sort.Slice(modelItems, func(i, j int) bool {
		if modelItems[i].GroupName == modelItems[j].GroupName {
			return modelItems[i].Name < modelItems[j].Name
		}
		return modelItems[i].GroupName < modelItems[j].GroupName
	})

	return PluginManagementEnums{
		Types: []string{"builtin", "custom"},
		Placements: []EnumOption{
			{Label: "资源详情动作区", Value: pluginx.PlacementResourceDetailActions},
		},
		UIs: []EnumOption{
			{Label: "在线终端", Value: pluginx.UIBuiltinTerminal},
			{Label: "文件管理", Value: pluginx.UIBuiltinSFTP},
		},
		Directions: []EnumOption{
			{Label: "源端", Value: pluginx.DirectionToSource},
			{Label: "目标端", Value: pluginx.DirectionToTarget},
		},
		RelationTypes: []EnumOption{
			{Label: "默认关系", Value: pluginx.RelationTypeDefault},
			{Label: "分组关系", Value: pluginx.RelationTypeGroup},
			{Label: "归属关系", Value: pluginx.RelationTypeBelong},
			{Label: "运行关系", Value: pluginx.RelationTypeRun},
		},
		Cardinalities: []EnumOption{
			{Label: "单个", Value: pluginx.CardinalityOne},
			{Label: "多个", Value: pluginx.CardinalityMany},
		},
		Mappings: []EnumOption{
			{Label: "一对一", Value: pluginx.MappingOneToOne},
			{Label: "一对多", Value: pluginx.MappingOneToMany},
			{Label: "多对多", Value: pluginx.MappingManyToMany},
		},
		Models: modelItems,
	}, nil
}



func (s *service) DeletePlugin(ctx context.Context, uid string) error {
	uid = strings.TrimSpace(uid)
	if uid == "" {
		return fmt.Errorf("plugin uid 不能为空")
	}

	plugin, err := s.repo.GetPlugin(ctx, uid)
	if err != nil {
		return err
	}
	if strings.EqualFold(plugin.Type, "builtin") {
		return errs.ValidationError.WithMsg("内置插件不允许删除")
	}
	return s.repo.DeletePlugin(ctx, uid)
}

type modelMeta struct {
	Name      string
	GroupName string
	Icon      string
	Builtin   bool
}

func buildBoundModels(bindings []domain.PluginBinding, modelMeta map[string]modelMeta) []PluginBoundModel {
	seen := map[string]struct{}{}
	items := make([]PluginBoundModel, 0, len(bindings))
	for _, binding := range bindings {
		if _, ok := seen[binding.ModelUID]; ok {
			continue
		}
		seen[binding.ModelUID] = struct{}{}
		meta := modelMeta[binding.ModelUID]
		items = append(items, PluginBoundModel{
			UID:       binding.ModelUID,
			Name:      meta.Name,
			GroupName: meta.GroupName,
			Icon:      meta.Icon,
			Builtin:   meta.Builtin,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].GroupName == items[j].GroupName {
			return items[i].Name < items[j].Name
		}
		return items[i].GroupName < items[j].GroupName
	})
	return items
}

func (s *service) loadModelMetaByUID(ctx context.Context, modelUIDs []string) (map[string]modelMeta, error) {
	if len(modelUIDs) == 0 {
		return map[string]modelMeta{}, nil
	}

	models, err := s.models.GetByUids(ctx, modelUIDs)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return map[string]modelMeta{}, nil
		}
		return nil, err
	}

	groupMap, err := s.loadModelGroupNameMap(ctx, models)
	if err != nil {
		return nil, err
	}

	result := make(map[string]modelMeta, len(models))
	for _, item := range models {
		result[item.UID] = modelMeta{
			Name:      item.Name,
			GroupName: groupMap[item.GroupId],
			Icon:      item.Icon,
			Builtin:   item.Builtin,
		}
	}
	return result, nil
}

func (s *service) loadModelGroupNameMap(ctx context.Context, models []domain.Model) (map[int64]string, error) {
	groupIDs := lo.Uniq(lo.FilterMap(models, func(item domain.Model, _ int) (int64, bool) {
		return item.GroupId, item.GroupId > 0
	}))
	if len(groupIDs) == 0 {
		return map[int64]string{}, nil
	}

	groups, _, err := s.modelGroups.List(ctx, 0, int64(len(groupIDs)+20))
	if err != nil {
		return nil, err
	}

	result := make(map[int64]string, len(groups))
	for _, group := range groups {
		result[group.ID] = group.Name
	}
	return result, nil
}
