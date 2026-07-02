package plugin

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/Duke1616/ecmdb/internal/domain"
	"github.com/Duke1616/ecmdb/internal/errs"
	_ "github.com/Duke1616/ecmdb/internal/plugin/builtin"
	"github.com/Duke1616/ecmdb/internal/repository"
	attribute "github.com/Duke1616/ecmdb/internal/service/attribute"
	model "github.com/Duke1616/ecmdb/internal/service/model"
	relation "github.com/Duke1616/ecmdb/internal/service/relation"
	resource "github.com/Duke1616/ecmdb/internal/service/resource"
	pluginx "github.com/Duke1616/ecmdb/pkg/plugin"
	"github.com/samber/lo"
)

type Service interface {
	// RegisterBuiltinPlugins 仅注册引导注册系统内置插件元数据。
	RegisterBuiltinPlugins(ctx context.Context) error

	// GetDefaultDefinition 返回插件默认定义草稿，供前端创建内置默认绑定时使用。
	GetDefaultDefinition(ctx context.Context, pluginID string) (pluginx.Definition, error)

	// SaveBindings 保存某个插件的绑定图；若绑定模型命中内置默认定义，则自动导入对应 Schema。
	SaveBindings(ctx context.Context, req domain.SavePluginBindings) error

	// ToggleBindingStatus 切换单个绑定的启停状态，并返回切换后的状态。
	ToggleBindingStatus(ctx context.Context, uid string) (bool, error)

	// DeleteBinding 删除单个插件绑定。
	DeleteBinding(ctx context.Context, uid string) error

	// ListPlugins 查询插件目录。
	ListPlugins(ctx context.Context) ([]domain.PluginListItem, error)

	// GetPluginDetail 查询插件详情。
	GetPluginDetail(ctx context.Context, uid string) (domain.PluginDetail, error)

	// ListEnums 查询插件管理所需枚举。
	ListEnums(ctx context.Context) (domain.PluginManagementEnums, error)

	// ListResourceActionsBatch 批量查询多个资源可以使用的插件动作。
	ListResourceActionsBatch(ctx context.Context, resourceIDs []int64) ([]pluginx.ResourceActions, error)

	// ResolveAction 解析插件动作需要的 UI 和输入数据。
	ResolveAction(ctx context.Context, req pluginx.ResolveRequest) (pluginx.ResolveResult, error)

	// ResolveActionContext 解析插件动作运行时上下文，供内置后端能力直接复用。
	ResolveActionContext(ctx context.Context, req pluginx.ResolveRequest) (pluginx.ActionContext, error)
}

type service struct {
	repo           repository.PluginRepository
	resolver       *inputResolver
	models         model.Service
	modelGroups    model.MGService
	attributes     attribute.Service
	relationTypes  relation.RelationTypeService
	modelRelations relation.RelationModelService
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
	resourceSvc resource.Service,
	relationSvc relation.RelationResourceService,
	modelSvc model.Service,
	modelGroupSvc model.MGService,
	attributeSvc attribute.Service,
	relationTypeSvc relation.RelationTypeService,
	relationModelSvc relation.RelationModelService,
) Service {
	return &service{
		repo:           repo,
		models:         modelSvc,
		modelGroups:    modelGroupSvc,
		attributes:     attributeSvc,
		relationTypes:  relationTypeSvc,
		modelRelations: relationModelSvc,
		resolver: newInputResolver(
			resourceSvc,
			relationSvc,
		),
	}
}

func (s *service) RegisterBuiltinPlugins(ctx context.Context) error {
	for _, builtin := range pluginx.Builtins() {
		def := builtin.Definition()
		if err := s.syncBuiltinPlugin(ctx, def.Plugin); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) GetDefaultDefinition(ctx context.Context, pluginID string) (pluginx.Definition, error) {
	_ = ctx
	pluginID = strings.TrimSpace(pluginID)
	if pluginID == "" {
		return pluginx.Definition{}, fmt.Errorf("plugin_id 不能为空")
	}

	builtin, ok := pluginx.FindBuiltin(pluginID)
	if !ok {
		return pluginx.Definition{}, fmt.Errorf("未找到对应的内置插件定义: %s", pluginID)
	}
	return builtin.Definition(), nil
}

func (s *service) SaveBindings(ctx context.Context, req domain.SavePluginBindings) error {
	plan, err := s.buildBindingSavePlan(ctx, req)
	if err != nil {
		return err
	}

	if err = s.importBindingSchema(ctx, plan.pluginID, plan.bindings); err != nil {
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

	modelMeta, err := s.loadModelMetaByUID(ctx, lo.Uniq(lo.Map(bindings, func(item domain.PluginBinding, _ int) string {
		return item.ModelUID
	})))
	if err != nil {
		return nil, err
	}

	bindingsByPluginID := lo.GroupBy(bindings, func(item domain.PluginBinding) string {
		return item.PluginID
	})

	items := make([]domain.PluginListItem, 0, len(plugins))
	for _, item := range plugins {
		pluginBindings := bindingsByPluginID[item.UID]
		items = append(items, domain.PluginListItem{
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

	modelMeta, err := s.loadModelMetaByUID(ctx, lo.Uniq(lo.Map(bindings, func(item domain.PluginBinding, _ int) string {
		return item.ModelUID
	})))
	if err != nil {
		return domain.PluginDetail{}, err
	}

	details := make([]domain.PluginBindingDetail, 0, len(bindings))
	for _, binding := range bindings {
		normalized, err := pluginx.PrepareBinding(binding)
		if err != nil {
			return domain.PluginDetail{}, err
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

	return domain.PluginDetail{
		Plugin:   plugin,
		Bindings: details,
	}, nil
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
	sort.Slice(modelItems, func(i, j int) bool {
		if modelItems[i].GroupName == modelItems[j].GroupName {
			return modelItems[i].Name < modelItems[j].Name
		}
		return modelItems[i].GroupName < modelItems[j].GroupName
	})

	return domain.PluginManagementEnums{
		Types: []string{"builtin", "custom"},
		Placements: []domain.EnumOption{
			{Label: "资源详情动作区", Value: pluginx.PlacementResourceDetailActions},
		},
		UIs: []domain.EnumOption{
			{Label: "在线终端", Value: pluginx.UIBuiltinTerminal},
			{Label: "文件管理", Value: pluginx.UIBuiltinSFTP},
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

	normalizedIDs := make([]int64, 0, len(resourceIDs))
	for _, resourceID := range resourceIDs {
		if err := pluginx.ValidateResourceID(resourceID); err != nil {
			return nil, err
		}
		normalizedIDs = append(normalizedIDs, resourceID)
	}

	resources, err := s.resolver.resources.ListResourceByIds(ctx, []string{}, lo.Uniq(normalizedIDs))
	if err != nil {
		return nil, err
	}

	resourceMap := lo.SliceToMap(resources, func(item domain.Resource) (int64, domain.Resource) {
		return item.ID, item
	})
	bindingCache := make(map[string][]domain.PluginBinding)
	pluginCache := make(map[string]domain.Plugin)

	results := make([]pluginx.ResourceActions, 0, len(normalizedIDs))
	for _, resourceID := range normalizedIDs {
		resource, ok := resourceMap[resourceID]
		if !ok {
			return nil, fmt.Errorf("资源不存在: %d", resourceID)
		}

		bindings, ok := bindingCache[resource.ModelUID]
		if !ok {
			bindings, err = s.repo.ListEnabledBindingsByModelUID(ctx, resource.ModelUID)
			if err != nil {
				return nil, err
			}
			bindingCache[resource.ModelUID] = bindings
		}

		actions, err := s.listActionsByBindingsWithCache(ctx, bindings, func(binding domain.PluginBinding) (bool, error) {
			return s.resolver.bindingSatisfied(ctx, resource, binding)
		}, pluginCache)
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
	prepared, err := pluginx.PrepareBinding(b)
	if err != nil {
		return err
	}
	return s.repo.UpsertBinding(ctx, prepared)
}

func (s *service) loadPlugin(ctx context.Context, pluginID string) (domain.Plugin, error) {
	return s.repo.GetPlugin(ctx, pluginID)
}

func (s *service) syncBuiltinPlugin(ctx context.Context, next pluginx.Plugin) error {
	plugin, err := s.keepPluginRuntimeState(ctx, next)
	if err != nil {
		return err
	}
	return s.upsertPlugin(ctx, plugin)
}

func (s *service) keepPluginRuntimeState(ctx context.Context, next pluginx.Plugin) (pluginx.Plugin, error) {
	existing, err := s.repo.GetPlugin(ctx, next.UID)
	switch {
	case err == nil:
		next.ID = existing.ID
		return next, nil
	case errors.Is(err, errs.ErrNotFound):
		return next, nil
	default:
		return pluginx.Plugin{}, err
	}
}

func (s *service) buildBindingSavePlan(ctx context.Context, req domain.SavePluginBindings) (bindingSavePlan, error) {
	pluginID := strings.TrimSpace(req.PluginID)
	if pluginID == "" {
		return bindingSavePlan{}, fmt.Errorf("plugin_id 不能为空")
	}

	if _, err := s.loadPlugin(ctx, pluginID); err != nil {
		return bindingSavePlan{}, err
	}

	bindings, err := pluginx.PrepareBindings(pluginID, req.Bindings)
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

func (s *service) importBindingSchema(ctx context.Context, pluginID string, bindings []pluginx.Binding) error {
	builtin, ok := pluginx.FindBuiltin(pluginID)
	if !ok {
		return nil
	}

	schema, err := builtin.SchemaForBindings(bindings)
	if err != nil {
		return err
	}
	if isEmptySchema(schema) {
		return nil
	}
	return s.importSchema(ctx, schema)
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
			return s.resolver.bindingSatisfied(ctx, primary, binding)
		},
	)
}

func (s *service) listActionsByBindingsWithCache(
	ctx context.Context,
	bindings []domain.PluginBinding,
	match func(binding domain.PluginBinding) (bool, error),
	pluginCache map[string]domain.Plugin,
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

		plugin, err := s.loadCachedPlugin(ctx, binding.PluginID, pluginCache)
		if err != nil {
			return nil, err
		}

		actions = append(actions, plugin.ResourceActions()...)
	}
	return actions, nil
}

func (s *service) listActionsByBindings(
	ctx context.Context,
	bindings []domain.PluginBinding,
	match func(binding domain.PluginBinding) (bool, error),
) ([]pluginx.ResourceAction, error) {
	return s.listActionsByBindingsWithCache(ctx, bindings, match, make(map[string]domain.Plugin, len(bindings)))
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
	if err := pluginx.ValidateResolveRequest(req); err != nil {
		return actionTarget{}, err
	}

	resource, err := s.resolver.loadResource(ctx, req.ResourceID, nil)
	if err != nil {
		return actionTarget{}, err
	}

	binding, err := s.findBinding(ctx, resource.ModelUID, req.PluginID)
	if err != nil {
		return actionTarget{}, err
	}

	plugin, err := s.loadPlugin(ctx, binding.PluginID)
	if err != nil {
		return actionTarget{}, err
	}

	action, ok := plugin.FindAction(req.Action)
	if !ok {
		return actionTarget{}, fmt.Errorf("插件动作不存在: %s", req.Action)
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
	inputs, err := s.resolver.resolve(ctx, target.resource, target.binding.Graph)
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

func (s *service) findBinding(ctx context.Context, modelUID string, pluginID string) (domain.PluginBinding, error) {
	bindings, err := s.repo.ListEnabledBindingsByModelUID(ctx, modelUID)
	if err != nil {
		return domain.PluginBinding{}, err
	}

	binding, ok := lo.Find(bindings, func(binding domain.PluginBinding) bool {
		return binding.PluginID == pluginID
	})
	if !ok {
		return domain.PluginBinding{}, fmt.Errorf("插件绑定不存在")
	}
	return binding, nil
}

func resolveResult(actionCtx pluginx.ActionContext) pluginx.ResolveResult {
	return pluginx.ResolveResult{
		UI:         actionCtx.Action.UI,
		PluginID:   actionCtx.Plugin.UID,
		PluginName: actionCtx.Plugin.Name,
		Action:     actionCtx.Action.Action,
		ResourceID: actionCtx.ResourceID,
		Inputs:     actionCtx.Inputs,
		Params:     actionCtx.Params,
		Meta:       actionCtx.Action.Meta,
	}
}

func (s *service) importSchema(ctx context.Context, schema pluginx.Schema) error {
	if err := s.importModels(ctx, schema.ModelGroups, schema.Models); err != nil {
		return err
	}
	if err := s.importRelationTypes(ctx, schema.RelationTypes); err != nil {
		return err
	}
	return s.importModelRelations(ctx, schema.ModelRelations)
}

func (s *service) importModels(ctx context.Context, modelGroups []pluginx.ModelGroupSpec, models []pluginx.ModelSpec) error {
	groups, err := s.ensureModelGroups(ctx, modelGroups, models)
	if err != nil {
		return err
	}

	for _, model := range models {
		groupID := groups[model.GroupName]
		if err = s.ensureModel(ctx, model, groupID); err != nil {
			return err
		}
		if err = s.ensureAttributes(ctx, model); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) ensureModelGroups(
	ctx context.Context,
	modelGroups []pluginx.ModelGroupSpec,
	models []pluginx.ModelSpec,
) (map[string]int64, error) {
	names := lo.Map(modelGroups, func(group pluginx.ModelGroupSpec, _ int) string {
		return group.Name
	})
	names = append(names, lo.FilterMap(models, func(model pluginx.ModelSpec, _ int) (string, bool) {
		return model.GroupName, model.GroupName != ""
	})...)
	names = lo.Uniq(lo.Filter(names, func(name string, _ int) bool {
		return name != ""
	}))
	if len(names) == 0 {
		return map[string]int64{}, nil
	}

	existing, err := s.modelGroups.GetByNames(ctx, names)
	if err != nil {
		return nil, err
	}

	byName := lo.SliceToMap(existing, func(group domain.ModelGroup) (string, domain.ModelGroup) {
		return group.Name, group
	})
	missing := lo.FilterMap(names, func(name string, _ int) (domain.ModelGroup, bool) {
		_, ok := byName[name]
		return domain.ModelGroup{Name: name}, !ok
	})
	if len(missing) > 0 {
		created, err := s.modelGroups.BatchCreate(ctx, missing)
		if err != nil {
			return nil, fmt.Errorf("创建插件模型分组失败: %w", err)
		}
		for _, group := range created {
			byName[group.Name] = group
		}
	}

	return lo.MapValues(byName, func(group domain.ModelGroup, _ string) int64 {
		return group.ID
	}), nil
}

func (s *service) ensureModel(ctx context.Context, spec pluginx.ModelSpec, groupID int64) error {
	if spec.UID == "" {
		return fmt.Errorf("插件模型 UID 不能为空")
	}
	if spec.Name == "" {
		return fmt.Errorf("插件模型名称不能为空: %s", spec.UID)
	}

	_, err := s.models.GetByUid(ctx, spec.UID)
	switch {
	case err == nil:
		return nil
	case errors.Is(err, errs.ErrNotFound):
		_, err = s.models.Create(ctx, domain.Model{
			UID:     spec.UID,
			Name:    spec.Name,
			Icon:    spec.Icon,
			GroupId: groupID,
			Builtin: spec.Builtin,
		})
		if err != nil {
			return fmt.Errorf("创建插件模型失败 %s: %w", spec.UID, err)
		}
		return nil
	default:
		return err
	}
}

func (s *service) ensureAttributes(ctx context.Context, model pluginx.ModelSpec) error {
	if len(model.AttributeGroups) == 0 {
		return nil
	}

	groups, err := s.ensureAttributeGroups(ctx, model)
	if err != nil {
		return err
	}
	return s.ensureAttributeFields(ctx, model, groups)
}

func (s *service) ensureAttributeGroups(ctx context.Context, model pluginx.ModelSpec) (map[string]int64, error) {
	existing, err := s.attributes.ListAttributeGroup(ctx, model.UID)
	if err != nil {
		return nil, err
	}

	byName := lo.SliceToMap(existing, func(group domain.AttributeGroup) (string, domain.AttributeGroup) {
		return group.Name, group
	})
	missing := lo.FilterMap(model.AttributeGroups, func(group pluginx.AttributeGroup, _ int) (domain.AttributeGroup, bool) {
		_, ok := byName[group.Name]
		return domain.AttributeGroup{
			Name:     group.Name,
			ModelUid: model.UID,
			SortKey:  group.Index,
		}, group.Name != "" && !ok
	})
	if len(missing) > 0 {
		created, err := s.attributes.BatchCreateAttributeGroup(ctx, missing)
		if err != nil {
			return nil, fmt.Errorf("创建插件模型属性分组失败 %s: %w", model.UID, err)
		}
		for _, group := range created {
			byName[group.Name] = group
		}
	}

	return lo.MapValues(byName, func(group domain.AttributeGroup, _ string) int64 {
		return group.ID
	}), nil
}

func (s *service) ensureAttributeFields(ctx context.Context, model pluginx.ModelSpec, groups map[string]int64) error {
	existing, _, err := s.attributes.ListAttributes(ctx, model.UID)
	if err != nil {
		return err
	}
	existingFields := lo.SliceToMap(existing, func(attr domain.Attribute) (string, struct{}) {
		return attr.FieldUid, struct{}{}
	})

	fields := make([]domain.Attribute, 0)
	for _, group := range model.AttributeGroups {
		groupID := groups[group.Name]
		for _, field := range group.Fields {
			if field.UID == "" {
				return fmt.Errorf("插件模型字段 UID 不能为空: %s.%s", model.UID, group.Name)
			}
			if _, ok := existingFields[field.UID]; ok {
				continue
			}
			fields = append(fields, domain.Attribute{
				GroupId:   groupID,
				ModelUid:  model.UID,
				FieldUid:  field.UID,
				FieldName: field.Name,
				FieldType: field.Type,
				Required:  field.Required,
				Display:   field.Display,
				Secure:    field.Secure,
				Index:     field.Index,
				SortKey:   field.Index,
				Option:    field.Option,
				Builtin:   field.Builtin,
			})
		}
	}
	if len(fields) == 0 {
		return nil
	}
	return s.attributes.BatchCreateAttribute(ctx, fields)
}

func (s *service) importRelationTypes(ctx context.Context, relationTypes []pluginx.RelationType) error {
	if len(relationTypes) == 0 {
		return nil
	}

	uids := lo.Map(relationTypes, func(relationType pluginx.RelationType, _ int) string {
		return relationType.UID
	})
	existing, err := s.relationTypes.GetByUids(ctx, uids)
	if err != nil {
		return err
	}
	existingUIDs := lo.SliceToMap(existing, func(relationType domain.RelationType) (string, struct{}) {
		return relationType.UID, struct{}{}
	})

	missing := lo.FilterMap(relationTypes, func(relationType pluginx.RelationType, _ int) (domain.RelationType, bool) {
		_, ok := existingUIDs[relationType.UID]
		return domain.RelationType{
			UID:            relationType.UID,
			Name:           relationType.Name,
			SourceDescribe: relationType.SourceDescribe,
			TargetDescribe: relationType.TargetDescribe,
		}, relationType.UID != "" && !ok
	})
	return s.relationTypes.BatchCreate(ctx, missing)
}

func (s *service) importModelRelations(ctx context.Context, relations []pluginx.ModelRelation) error {
	if len(relations) == 0 {
		return nil
	}

	modelRelations := lo.Map(relations, func(relation pluginx.ModelRelation, _ int) domain.ModelRelation {
		return domain.ModelRelation{
			SourceModelUID:  relation.SourceModelUID,
			TargetModelUID:  relation.TargetModelUID,
			RelationTypeUID: relation.RelationTypeUID,
			Mapping:         relation.Mapping,
		}
	})

	for i := range modelRelations {
		if err := modelRelations[i].Validate(); err != nil {
			return err
		}
	}

	names := lo.Map(modelRelations, func(relation domain.ModelRelation, _ int) string {
		return relation.RelationName
	})
	existing, err := s.modelRelations.GetByRelationNames(ctx, names)
	if err != nil {
		return err
	}
	existingByName := lo.SliceToMap(existing, func(relation domain.ModelRelation) (string, domain.ModelRelation) {
		return relation.RelationName, relation
	})

	missing := make([]domain.ModelRelation, 0)
	for _, relation := range modelRelations {
		existingRelation, ok := existingByName[relation.RelationName]
		if !ok {
			missing = append(missing, relation)
			continue
		}
		if !modelRelationChanged(existingRelation, relation) {
			continue
		}

		relation.ID = existingRelation.ID
		if _, err = s.modelRelations.UpdateModelRelation(ctx, relation); err != nil {
			return err
		}
	}
	return s.modelRelations.BatchCreate(ctx, missing)
}

func modelRelationChanged(current domain.ModelRelation, next domain.ModelRelation) bool {
	return current.SourceModelUID != next.SourceModelUID ||
		current.TargetModelUID != next.TargetModelUID ||
		current.RelationTypeUID != next.RelationTypeUID ||
		current.Mapping != next.Mapping
}

func buildBoundModels(bindings []domain.PluginBinding, modelMeta map[string]modelMeta) []domain.PluginBoundModel {
	seen := map[string]struct{}{}
	items := make([]domain.PluginBoundModel, 0, len(bindings))
	for _, binding := range bindings {
		if _, ok := seen[binding.ModelUID]; ok {
			continue
		}
		seen[binding.ModelUID] = struct{}{}
		meta := modelMeta[binding.ModelUID]
		items = append(items, domain.PluginBoundModel{
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

	groups, err := s.modelGroups.GetByIDs(ctx, groupIDs)
	if err != nil {
		return nil, err
	}

	result := make(map[int64]string, len(groups))
	for _, group := range groups {
		result[group.ID] = group.Name
	}
	return result, nil
}

func isEmptySchema(schema pluginx.Schema) bool {
	return len(schema.Models) == 0 &&
		len(schema.ModelGroups) == 0 &&
		len(schema.RelationTypes) == 0 &&
		len(schema.ModelRelations) == 0
}
