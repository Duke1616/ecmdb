package plugin

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/Duke1616/ecmdb/internal/domain"
	"github.com/Duke1616/ecmdb/internal/errs"
	"github.com/Duke1616/ecmdb/internal/repository"
	model "github.com/Duke1616/ecmdb/internal/service/model"
	coreplugin "github.com/Duke1616/ecmdb/pkg/plugin"
	"github.com/Duke1616/ecmdb/pkg/plugin/graph"
	pluginx "github.com/Duke1616/ecmdb/pkg/plugin/types"
	"github.com/samber/lo"
)

//go:generate mockgen -source=./service.go -destination=./mocks/service.mock.go -package=pluginmocks -typed Service
//go:generate mockgen -package=pluginmocks -destination=./mocks/repository.mock.go -typed github.com/Duke1616/ecmdb/internal/repository PluginRepository
type Service interface {
	// ImportDefinition 导入外部插件定义。
	ImportDefinition(ctx context.Context, def coreplugin.Definition) error

	// GetDefaultDefinition 返回插件默认定义草稿，供前端创建内置默认绑定时使用。
	GetDefaultDefinition(ctx context.Context, pluginID string) (coreplugin.Definition, error)

	// SaveBindings 保存某个插件的绑定图；若绑定模型命中内置默认定义，则自动导入对应 Schema。
	SaveBindings(ctx context.Context, req domain.SavePluginBindings) error

	// ToggleBindingStatus 切换单个绑定的启停状态，并返回切换后的状态。
	ToggleBindingStatus(ctx context.Context, uid string) (bool, error)

	// DeleteBinding 删除单个插件绑定。
	DeleteBinding(ctx context.Context, uid string) error

	// ListPlugins 获取插件管理列表。
	ListPlugins(ctx context.Context) ([]domain.PluginListItem, error)

	// GetPluginDetail 获取插件详情与绑定列表。
	GetPluginDetail(ctx context.Context, uid string) (domain.PluginDetail, error)

	// ListEnums 获取插件管理所需的下拉字典选项。
	ListEnums(ctx context.Context) (domain.PluginManagementEnums, error)

	// ListResourceActionsBatch 批量获取多个资源当前可触发的插件动作。
	ListResourceActionsBatch(ctx context.Context, resourceIDs []int64) ([]pluginx.ResourceActions, error)

	// ResolveAction 根据资源与插件动作，解析动作运行所需上下文。
	ResolveAction(ctx context.Context, req pluginx.ResolveRequest) (pluginx.ResolveResult, error)

	// ResolveActionContext 根据资源与插件动作，解析完整的动作运行上下文。
	ResolveActionContext(ctx context.Context, req pluginx.ResolveRequest) (pluginx.ActionContext, error)

	// GetActionRuntime 获取插件动作运行时定义。
	GetActionRuntime(ctx context.Context, pluginID, action string) (domain.Plugin, pluginx.ActionSpec, error)
}

type service struct {
	repo        repository.PluginRepository
	resolver    IInputResolver
	importer    ISchemaImporter
	models      model.Service
	modelGroups model.MGService
}

type actionTarget struct {
	resource domain.Resource
	binding  domain.PluginBinding
	plugin   domain.Plugin
	action   domain.PluginActionSpec
}

type bindingSavePlan struct {
	pluginID string
	bindings []pluginx.Binding
}

type modelMeta struct {
	Name      string
	GroupName string
	Icon      string
	Builtin   bool
}

func NewService(
	repo repository.PluginRepository,
	resolver IInputResolver,
	importer ISchemaImporter,
	modelSvc model.Service,
	modelGroupSvc model.MGService,
) Service {
	return &service{
		repo:        repo,
		resolver:    resolver,
		importer:    importer,
		models:      modelSvc,
		modelGroups: modelGroupSvc,
	}
}

func (s *service) ImportDefinition(ctx context.Context, def coreplugin.Definition) error {
	return s.upsertPlugin(ctx, def.Plugin)
}

func (s *service) GetDefaultDefinition(ctx context.Context, pluginID string) (coreplugin.Definition, error) {
	pluginID = strings.TrimSpace(pluginID)
	if pluginID == "" {
		return coreplugin.Definition{}, fmt.Errorf("plugin_id 不能为空")
	}

	plugin, err := s.loadPlugin(ctx, pluginID)
	if err != nil {
		return coreplugin.Definition{}, err
	}

	if def, err := loadRuntimeDefaultDefinition(ctx, plugin); err == nil {
		return def, nil
	} else {
		return coreplugin.Definition{}, fmt.Errorf("插件默认定义获取失败，请确认插件运行时可访问: %s: %w", pluginID, err)
	}
}

func loadRuntimeDefaultDefinition(ctx context.Context, plugin domain.Plugin) (coreplugin.Definition, error) {
	runtime, ok := plugin.Runtime()
	if !ok || strings.TrimSpace(runtime.Upstream) == "" {
		return coreplugin.Definition{}, fmt.Errorf("插件 runtime upstream 为空: %s", plugin.UID)
	}

	def, err := coreplugin.FetchDefinition(ctx, runtime.Upstream)
	if err != nil {
		return coreplugin.Definition{}, err
	}
	if strings.TrimSpace(def.Plugin.UID) != strings.TrimSpace(plugin.UID) {
		return coreplugin.Definition{}, fmt.Errorf("插件自描述 UID 不匹配: got %s, want %s", def.Plugin.UID, plugin.UID)
	}

	bindings, err := prepareDefinitionBindings(def.Plugin.UID, def.Bindings)
	if err != nil {
		return coreplugin.Definition{}, err
	}
	def.Bindings = bindings
	return def, nil
}

func (s *service) SaveBindings(ctx context.Context, req domain.SavePluginBindings) error {
	plan, err := s.buildBindingSavePlan(ctx, req)
	if err != nil {
		return err
	}

	if err = s.importRuntimePluginSchema(ctx, plan.pluginID, plan.bindings); err != nil {
		return err
	}

	return s.savePreparedBindings(ctx, plan.bindings)
}

func (s *service) ToggleBindingStatus(ctx context.Context, uid string) (bool, error) {
	uid = strings.TrimSpace(uid)
	if uid == "" {
		return false, fmt.Errorf("binding uid 不能为空")
	}

	binding, err := s.repo.GetBinding(ctx, uid)
	if err != nil {
		return false, err
	}

	nextEnabled := !binding.Enabled
	if err = s.repo.UpdateBindingEnabled(ctx, uid, nextEnabled); err != nil {
		return false, err
	}
	return nextEnabled, nil
}

func (s *service) DeleteBinding(ctx context.Context, uid string) error {
	uid = strings.TrimSpace(uid)
	if uid == "" {
		return fmt.Errorf("binding uid 不能为空")
	}

	return s.repo.DeleteBinding(ctx, uid)
}

func (s *service) ListPlugins(ctx context.Context) ([]domain.PluginListItem, error) {
	plugins, err := s.repo.ListPlugins(ctx)
	if err != nil {
		return nil, err
	}
	if len(plugins) == 0 {
		return []domain.PluginListItem{}, nil
	}

	pluginIDs := lo.Map(plugins, func(item domain.Plugin, _ int) string {
		return item.UID
	})
	bindings, err := s.repo.ListBindingsByPluginIDs(ctx, pluginIDs)
	if err != nil {
		return nil, err
	}

	modelMeta, err := s.loadBindingsModelMeta(ctx, bindings)
	if err != nil {
		return nil, err
	}

	bindingsByPluginID := lo.GroupBy(bindings, func(item domain.PluginBinding) string {
		return item.PluginID
	})

	return lo.Map(plugins, func(item domain.Plugin, _ int) domain.PluginListItem {
		pluginBindings := bindingsByPluginID[item.UID]
		return domain.PluginListItem{
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
		}
	}), nil
}

func (s *service) GetPluginDetail(ctx context.Context, uid string) (domain.PluginDetail, error) {
	uid = strings.TrimSpace(uid)
	if uid == "" {
		return domain.PluginDetail{}, fmt.Errorf("plugin uid 不能为空")
	}

	plugin, err := s.repo.GetPlugin(ctx, uid)
	if err != nil {
		return domain.PluginDetail{}, err
	}

	bindings, err := s.repo.ListBindingsByPluginID(ctx, uid)
	if err != nil {
		return domain.PluginDetail{}, err
	}

	modelMeta, err := s.loadBindingsModelMeta(ctx, bindings)
	if err != nil {
		return domain.PluginDetail{}, err
	}

	details, err := buildPluginBindingDetails(bindings, modelMeta)
	if err != nil {
		return domain.PluginDetail{}, err
	}

	return domain.PluginDetail{
		Plugin:   plugin,
		Bindings: details,
	}, nil
}

func buildPluginBindingDetails(bindings []domain.PluginBinding, modelMeta map[string]modelMeta) ([]domain.PluginBindingDetail, error) {
	details := make([]domain.PluginBindingDetail, 0, len(bindings))
	for _, binding := range bindings {
		normalized, err := graph.PrepareBinding(binding)
		if err != nil {
			return nil, err
		}
		meta := modelMeta[normalized.ModelUID]
		details = append(details, domain.PluginBindingDetail{
			ID:        normalized.ID,
			UID:       normalized.UID,
			PluginID:  normalized.PluginID,
			ModelUID:  normalized.ModelUID,
			ModelName: meta.Name,
			GroupName: meta.GroupName,
			ModelIcon: meta.Icon,
			Enabled:   normalized.Enabled,
			Graph:     normalized.Graph,
		})
	}
	return details, nil
}

func (s *service) ListEnums(ctx context.Context) (domain.PluginManagementEnums, error) {
	models, err := s.models.ListAll(ctx)
	if err != nil {
		return domain.PluginManagementEnums{}, err
	}

	groupMap, err := s.loadModelGroupNameMap(ctx, models)
	if err != nil {
		return domain.PluginManagementEnums{}, err
	}

	modelItems := lo.Map(models, func(item domain.Model, _ int) domain.PluginModelVM {
		return domain.PluginModelVM{
			UID:       item.UID,
			Name:      item.Name,
			GroupName: groupMap[item.GroupId],
			Icon:      item.Icon,
			Builtin:   item.Builtin,
		}
	})
	slices.SortFunc(modelItems, func(a, b domain.PluginModelVM) int {
		if a.GroupName == b.GroupName {
			return cmp.Compare(a.Name, b.Name)
		}
		return cmp.Compare(a.GroupName, b.GroupName)
	})

	return domain.PluginManagementEnums{
		Types: []string{"builtin", "custom"},
		Placements: []domain.EnumOption{
			{Label: "资源详情动作区", Value: pluginx.PlacementResourceDetailActions},
		},
		Directions: []domain.EnumOption{
			{Label: "源端", Value: pluginx.DirectionToSource},
			{Label: "目标端", Value: pluginx.DirectionToTarget},
		},
		RelationTypes: []domain.EnumOption{
			{Label: "默认关系", Value: pluginx.RelationTypeDefault},
			{Label: "分组关系", Value: pluginx.RelationTypeGroup},
			{Label: "归属关系", Value: pluginx.RelationTypeBelong},
			{Label: "运行关系", Value: pluginx.RelationTypeRun},
		},
		Cardinalities: []domain.EnumOption{
			{Label: "单个", Value: pluginx.CardinalityOne},
			{Label: "多个", Value: pluginx.CardinalityMany},
		},
		Mappings: []domain.EnumOption{
			{Label: "一对一", Value: pluginx.MappingOneToOne},
			{Label: "一对多", Value: pluginx.MappingOneToMany},
			{Label: "多对多", Value: pluginx.MappingManyToMany},
		},
		Models: modelItems,
	}, nil
}

func (s *service) ListResourceActionsBatch(ctx context.Context, resourceIDs []int64) ([]pluginx.ResourceActions, error) {
	if len(resourceIDs) == 0 {
		return []pluginx.ResourceActions{}, nil
	}

	if lo.SomeBy(resourceIDs, func(id int64) bool { return id <= 0 }) {
		return nil, fmt.Errorf("resource_id 参数错误")
	}

	resources, err := s.resolver.ListResourcesByIDs(ctx, []string{}, lo.Uniq(resourceIDs))
	if err != nil {
		return nil, err
	}

	resourceMap := lo.SliceToMap(resources, func(item domain.Resource) (int64, domain.Resource) {
		return item.ID, item
	})

	// 循环外按涉及的模型 UID 预加载全部启用的绑定，消灭循环内 IO
	modelUIDs := lo.Uniq(lo.Map(resources, func(r domain.Resource, _ int) string {
		return r.ModelUID
	}))
	bindingCache := make(map[string][]domain.PluginBinding, len(modelUIDs))
	for _, uid := range modelUIDs {
		bindings, err := s.repo.ListEnabledBindingsByModelUID(ctx, uid)
		if err != nil {
			return nil, err
		}
		bindingCache[uid] = bindings
	}

	pluginCache := make(map[string]domain.Plugin)
	actionCache := make(map[string][]pluginx.ResourceAction)

	results := make([]pluginx.ResourceActions, 0, len(resourceIDs))
	for _, resourceID := range resourceIDs {
		resource, ok := resourceMap[resourceID]
		if !ok {
			return nil, fmt.Errorf("资源不存在: %d", resourceID)
		}

		bindings := bindingCache[resource.ModelUID]
		actions, err := s.listActionsByBindingsWithCache(ctx, bindings, func(binding domain.PluginBinding) (bool, error) {
			return s.resolver.BindingSatisfied(ctx, resource, binding)
		}, pluginCache, actionCache)
		if err != nil {
			return nil, err
		}

		results = append(results, pluginx.ResourceActions{
			ResourceID: resourceID,
			Actions:    actions,
		})
	}

	return results, nil
}

func (s *service) ResolveAction(ctx context.Context, req pluginx.ResolveRequest) (pluginx.ResolveResult, error) {
	actionCtx, err := s.ResolveActionContext(ctx, req)
	if err != nil {
		return pluginx.ResolveResult{}, err
	}
	return resolveResult(actionCtx), nil
}

func (s *service) ResolveActionContext(ctx context.Context, req pluginx.ResolveRequest) (pluginx.ActionContext, error) {
	target, err := s.resolveActionTarget(ctx, req)
	if err != nil {
		return pluginx.ActionContext{}, err
	}

	inputs, err := s.resolveActionInputs(ctx, target)
	if err != nil {
		return pluginx.ActionContext{}, err
	}

	return pluginx.ActionContext{
		Plugin:     target.plugin,
		Binding:    target.binding,
		Action:     target.action,
		ResourceID: req.ResourceID,
		Inputs:     inputs,
		Params:     req.Params,
	}, nil
}

func (s *service) upsertPlugin(ctx context.Context, p pluginx.Plugin) error {
	if err := p.Validate(); err != nil {
		return err
	}
	return s.repo.UpsertPlugin(ctx, p)
}

func (s *service) upsertBinding(ctx context.Context, b pluginx.Binding) error {
	prepared, err := graph.PrepareBinding(b)
	if err != nil {
		return err
	}
	return s.repo.UpsertBinding(ctx, prepared)
}

func (s *service) loadPlugin(ctx context.Context, pluginID string) (domain.Plugin, error) {
	return s.repo.GetPlugin(ctx, pluginID)
}

func (s *service) buildBindingSavePlan(ctx context.Context, req domain.SavePluginBindings) (bindingSavePlan, error) {
	pluginID := strings.TrimSpace(req.PluginID)
	if pluginID == "" {
		return bindingSavePlan{}, fmt.Errorf("plugin_id 不能为空")
	}

	if _, err := s.loadPlugin(ctx, pluginID); err != nil {
		return bindingSavePlan{}, err
	}

	bindings, err := graph.PrepareBindings(pluginID, req.Bindings)
	if err != nil {
		return bindingSavePlan{}, err
	}

	return bindingSavePlan{
		pluginID: pluginID,
		bindings: bindings,
	}, nil
}

func (s *service) savePreparedBindings(ctx context.Context, bindings []pluginx.Binding) error {
	for _, binding := range bindings {
		if err := s.upsertBinding(ctx, binding); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) importRuntimePluginSchema(ctx context.Context, pluginID string, bindings []pluginx.Binding) error {
	plugin, err := s.loadPlugin(ctx, pluginID)
	if err != nil {
		return err
	}

	def, err := loadRuntimeDefaultDefinition(ctx, plugin)
	if err != nil {
		return fmt.Errorf("插件 schema 获取失败，请确认插件运行时可访问: %s: %w", pluginID, err)
	}
	schema, err := coreplugin.StaticBuiltin(def).SchemaForBindings(bindings)
	if err != nil {
		return err
	}
	if isEmptySchema(schema) {
		return nil
	}
	return s.importer.ImportSchema(ctx, schema)
}

func prepareDefinitionBindings(pluginID string, bindings []pluginx.Binding) ([]pluginx.Binding, error) {
	if len(bindings) == 0 {
		return []pluginx.Binding{}, nil
	}
	return graph.PrepareBindings(pluginID, bindings)
}

func (s *service) listActionsForResource(ctx context.Context, primary domain.Resource) ([]pluginx.ResourceAction, error) {
	bindings, err := s.repo.ListEnabledBindingsByModelUID(ctx, primary.ModelUID)
	if err != nil {
		return nil, err
	}

	return s.listActionsByBindings(
		ctx,
		bindings,
		func(binding domain.PluginBinding) (bool, error) {
			return s.resolver.BindingSatisfied(ctx, primary, binding)
		},
	)
}

func (s *service) listActionsByBindingsWithCache(
	ctx context.Context,
	bindings []domain.PluginBinding,
	match func(binding domain.PluginBinding) (bool, error),
	pluginCache map[string]domain.Plugin,
	actionCache map[string][]pluginx.ResourceAction,
) ([]pluginx.ResourceAction, error) {
	actions := make([]pluginx.ResourceAction, 0, len(bindings))
	for _, binding := range bindings {
		ok, err := match(binding)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		cachedActions, ok := actionCache[binding.PluginID]
		if !ok {
			plugin, err := s.loadCachedPlugin(ctx, binding.PluginID, pluginCache)
			if err != nil {
				return nil, err
			}
			cachedActions = plugin.ResourceActions()
			actionCache[binding.PluginID] = cachedActions
		}

		actions = append(actions, cachedActions...)
	}
	return actions, nil
}

func (s *service) listActionsByBindings(
	ctx context.Context,
	bindings []domain.PluginBinding,
	match func(binding domain.PluginBinding) (bool, error),
) ([]pluginx.ResourceAction, error) {
	return s.listActionsByBindingsWithCache(
		ctx,
		bindings,
		match,
		make(map[string]domain.Plugin, len(bindings)),
		make(map[string][]pluginx.ResourceAction, len(bindings)),
	)
}

func (s *service) loadCachedPlugin(
	ctx context.Context,
	pluginID string,
	pluginCache map[string]domain.Plugin,
) (domain.Plugin, error) {
	if plugin, ok := pluginCache[pluginID]; ok {
		return plugin, nil
	}

	plugin, err := s.loadPlugin(ctx, pluginID)
	if err != nil {
		return domain.Plugin{}, err
	}
	pluginCache[pluginID] = plugin
	return plugin, nil
}

func (s *service) resolveActionTarget(ctx context.Context, req pluginx.ResolveRequest) (actionTarget, error) {
	if err := req.Validate(); err != nil {
		return actionTarget{}, err
	}

	resource, err := s.resolver.LoadResource(ctx, req.ResourceID, nil)
	if err != nil {
		return actionTarget{}, err
	}

	plugin, err := s.loadPlugin(ctx, req.PluginID)
	if err != nil {
		return actionTarget{}, err
	}

	action, ok := plugin.FindAction(req.Action)
	if !ok {
		return actionTarget{}, fmt.Errorf("插件动作不存在: %s", req.Action)
	}
	if strings.TrimSpace(action.BindingUID) == "" {
		return actionTarget{}, fmt.Errorf("插件动作未声明 binding_uid: %s", req.Action)
	}

	binding, err := s.findBinding(ctx, resource.ModelUID, req.PluginID, action.BindingUID)
	if err != nil {
		return actionTarget{}, err
	}

	return actionTarget{
		resource: resource,
		binding:  binding,
		plugin:   plugin,
		action:   action,
	}, nil
}

func (s *service) resolveActionInputs(
	ctx context.Context,
	target actionTarget,
) (map[string]pluginx.ResolvedInput, error) {
	inputs, err := s.resolver.Resolve(ctx, target.resource, target.binding.Graph)
	if err == nil {
		return inputs, nil
	}
	if !errors.Is(err, errRequiredInputMissing) {
		return nil, err
	}
	return nil, errs.ValidationError.WithMsg(
		fmt.Sprintf("插件动作缺少必需输入: %s", missingInputMessage(err)),
	)
}

func (s *service) findBinding(ctx context.Context, modelUID string, pluginID string, bindingUID string) (domain.PluginBinding, error) {
	bindings, err := s.repo.ListEnabledBindingsByModelUID(ctx, modelUID)
	if err != nil {
		return domain.PluginBinding{}, err
	}

	binding, ok := lo.Find(bindings, func(binding domain.PluginBinding) bool {
		return binding.PluginID == pluginID && binding.UID == bindingUID
	})
	if !ok {
		return domain.PluginBinding{}, fmt.Errorf("插件绑定不存在: %s", bindingUID)
	}
	return binding, nil
}

func resolveResult(actionCtx pluginx.ActionContext) pluginx.ResolveResult {
	return pluginx.ResolveResult{
		PluginID:      actionCtx.Plugin.UID,
		PluginName:    actionCtx.Plugin.Name,
		PluginVersion: actionCtx.Plugin.Version,
		ActionName:    actionCtx.Action.Name,
		Action:        actionCtx.Action.Action,
		Permission:    actionCtx.Action.Permission,
		BindingUID:    actionCtx.Binding.UID,
		ModelUID:      actionCtx.Binding.ModelUID,
		ResourceID:    actionCtx.ResourceID,
		Inputs:        actionCtx.Inputs,
		Params:        actionCtx.Params,
		Runtime:       actionCtx.Action.Runtime,
		Meta:          actionCtx.Action.Meta,
	}
}

func buildBoundModels(bindings []domain.PluginBinding, modelMeta map[string]modelMeta) []domain.PluginBoundModel {
	uniqBindings := lo.UniqBy(bindings, func(b domain.PluginBinding) string {
		return b.ModelUID
	})
	items := lo.Map(uniqBindings, func(binding domain.PluginBinding, _ int) domain.PluginBoundModel {
		meta := modelMeta[binding.ModelUID]
		return domain.PluginBoundModel{
			UID:       binding.ModelUID,
			Name:      meta.Name,
			GroupName: meta.GroupName,
			Icon:      meta.Icon,
			Builtin:   meta.Builtin,
		}
	})
	slices.SortFunc(items, func(a, b domain.PluginBoundModel) int {
		if a.GroupName == b.GroupName {
			return cmp.Compare(a.Name, b.Name)
		}
		return cmp.Compare(a.GroupName, b.GroupName)
	})
	return items
}

func (s *service) loadBindingsModelMeta(ctx context.Context, bindings []domain.PluginBinding) (map[string]modelMeta, error) {
	modelUIDs := lo.Uniq(lo.Map(bindings, func(item domain.PluginBinding, _ int) string {
		return item.ModelUID
	}))
	return s.loadModelMetaByUID(ctx, modelUIDs)
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

	return lo.SliceToMap(models, func(item domain.Model) (string, modelMeta) {
		return item.UID, modelMeta{
			Name:      item.Name,
			GroupName: groupMap[item.GroupId],
			Icon:      item.Icon,
			Builtin:   item.Builtin,
		}
	}), nil
}

func (s *service) loadModelGroupNameMap(ctx context.Context, models []domain.Model) (map[int64]string, error) {
	groupIDs := lo.Uniq(lo.FilterMap(models, func(item domain.Model, _ int) (int64, bool) {
		return item.GroupId, item.GroupId > 0
	}))
	if len(groupIDs) == 0 {
		return map[int64]string{}, nil
	}

	groups, err := s.modelGroups.GetByIDs(ctx, groupIDs)
	if err != nil {
		return nil, err
	}

	return lo.SliceToMap(groups, func(g domain.ModelGroup) (int64, string) {
		return g.ID, g.Name
	}), nil
}

func (s *service) GetActionRuntime(ctx context.Context, pluginID, action string) (domain.Plugin, pluginx.ActionSpec, error) {
	pluginID = strings.TrimSpace(pluginID)
	action = strings.TrimSpace(action)
	if pluginID == "" || action == "" {
		return domain.Plugin{}, pluginx.ActionSpec{}, fmt.Errorf("plugin_id 或 action 不能为空")
	}

	p, err := s.loadPlugin(ctx, pluginID)
	if err != nil {
		return domain.Plugin{}, pluginx.ActionSpec{}, err
	}

	spec, ok := p.FindAction(action)
	if !ok {
		return domain.Plugin{}, pluginx.ActionSpec{}, fmt.Errorf("插件动作不存在: %s", action)
	}

	return p, spec, nil
}

