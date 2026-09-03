package plugin

import (
	"context"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/domain"
	relation "github.com/Duke1616/ecmdb/internal/service/relation"
	resource "github.com/Duke1616/ecmdb/internal/service/resource"
	"github.com/Duke1616/ecmdb/pkg/plugin/graph"
	pluginx "github.com/Duke1616/ecmdb/pkg/plugin/types"
	"github.com/samber/lo"
)

//go:generate mockgen -source=./resolver.go -destination=./mocks/resolver.mock.go -package=pluginmocks -typed IInputResolver
// IInputResolver 定义插件输入解析能力接口，负责资源动作匹配与关联拓扑图参数装配
type IInputResolver interface {
	// BindingSatisfied 在纯内存中判定资源是否满足绑定的入口过滤条件（零数据库 IO，毫秒级响应）
	BindingSatisfied(ctx context.Context, primary domain.Resource, binding domain.PluginBinding) (bool, error)

	// Resolve 递归执行绑定图的拓扑解析，装配插件动作触发所需的全部输入数据（包括凭证与上下游资产属性）
	Resolve(ctx context.Context, primary domain.Resource, bg *pluginx.BindingGraph) (map[string]pluginx.ResolvedInput, error)

	// LoadResource 查询指定资源并仅按需加载插件声明所需的字段，避免加载不必要的大字段
	LoadResource(ctx context.Context, resourceID int64, fields []string) (domain.Resource, error)

	// ListResourcesByIDs 批量查询多个关联资产，用于列表动作的批量预加载与按需裁剪
	ListResourcesByIDs(ctx context.Context, fields []string, ids []int64) ([]domain.Resource, error)
}

type inputResolver struct {
	resources resource.Service
	relations relation.RelationResourceService
}

func NewInputResolver(
	resources resource.Service,
	relations relation.RelationResourceService,
) IInputResolver {
	return &inputResolver{
		resources: resources,
		relations: relations,
	}
}

// BindingSatisfied 在纯内存中极轻量判定资源是否满足绑定的入口过滤条件（零数据库 IO，毫秒级响应）
func (r *inputResolver) BindingSatisfied(_ context.Context, primary domain.Resource, binding domain.PluginBinding) (bool, error) {
	if !binding.Enabled || binding.ModelUID != primary.ModelUID {
		return false, nil
	}

	if binding.Graph == nil {
		return true, nil
	}

	entryNode, ok := graph.GraphEntryNode(binding.Graph)
	if !ok || len(entryNode.Filters) == 0 {
		return true, nil
	}

	return resourceMatchesFilters(primary, entryNode.Filters), nil
}

func (r *inputResolver) Resolve(
	ctx context.Context,
	primary domain.Resource,
	bg *pluginx.BindingGraph,
) (map[string]pluginx.ResolvedInput, error) {
	specs, err := graph.CompileBindingGraph(bg)
	if err != nil {
		return nil, err
	}
	return r.resolveSpecsWithPath(ctx, primary, specs, true, "")
}

func (r *inputResolver) LoadResource(ctx context.Context, resourceID int64, fields []string) (domain.Resource, error) {
	if fields == nil {
		fields = []string{}
	}
	return r.resources.FindResourceById(ctx, fields, resourceID)
}

func (r *inputResolver) ListResourcesByIDs(ctx context.Context, fields []string, ids []int64) ([]domain.Resource, error) {
	return r.resources.ListResourceByIds(ctx, fields, ids)
}

func (r *inputResolver) resolveSpecsWithPath(
	ctx context.Context,
	base domain.Resource,
	specs []pluginx.ResourceSpec,
	topLevel bool,
	parentPath string,
) (map[string]pluginx.ResolvedInput, error) {
	if len(specs) == 0 {
		return nil, nil
	}

	resolved := make(map[string]pluginx.ResolvedInput, len(specs))
	for _, spec := range specs {
		path := joinSpecPath(parentPath, spec.Name)
		input, ok, err := r.resolveSpec(ctx, base, spec, topLevel, path)
		if err != nil {
			return nil, err
		}
		if !ok && spec.Required {
			return nil, newMissingInputError(missingSpecReason(path, spec, base))
		}
		resolved[spec.Name] = input
	}
	return resolved, nil
}

func (r *inputResolver) resolveSpec(
	ctx context.Context,
	base domain.Resource,
	spec pluginx.ResourceSpec,
	topLevel bool,
	path string,
) (pluginx.ResolvedInput, bool, error) {
	if topLevel && spec.RelationType == "" {
		return r.resolveCenterInput(ctx, base, spec, path)
	}

	resources, err := r.loadRelatedResources(ctx, base, spec)
	if err != nil {
		return emptyInput(spec), false, err
	}
	return r.resolveResourceList(ctx, spec, resources, path)
}

func (r *inputResolver) resolveCenterInput(
	ctx context.Context,
	base domain.Resource,
	spec pluginx.ResourceSpec,
	path string,
) (pluginx.ResolvedInput, bool, error) {
	if spec.ModelUID == "" {
		return emptyInput(spec), !spec.Required, nil
	}

	fields := specFields(spec)

	var (
		resource domain.Resource
		err      error
	)
	if base.ModelUID == spec.ModelUID && resourceHasFields(base, fields) {
		resource = base
	} else {
		resource, err = r.LoadResource(ctx, base.ID, fields)
		if err != nil {
			return emptyInput(spec), false, err
		}
	}
	if !resourceMatchesFilters(resource, spec.Filters) {
		return emptyInput(spec), !spec.Required, nil
	}

	item, ok, err := r.resolveResource(ctx, resource, spec, path)
	if err != nil || !ok {
		return emptyInput(spec), ok, err
	}
	resolved := emptyInput(spec)
	resolved.Resources = append(resolved.Resources, item)
	return resolved, true, nil
}

func (r *inputResolver) resolveResourceList(
	ctx context.Context,
	spec pluginx.ResourceSpec,
	resources []domain.Resource,
	path string,
) (pluginx.ResolvedInput, bool, error) {
	resolved := emptyInput(spec)
	if len(resources) == 0 {
		return resolved, !spec.Required, nil
	}

	for _, resource := range resources {
		item, ok, err := r.resolveResource(ctx, resource, spec, path)
		if err != nil {
			return resolved, false, err
		}
		if !ok {
			return resolved, false, nil
		}
		resolved.Resources = append(resolved.Resources, item)
	}
	return resolved, true, nil
}

func (r *inputResolver) resolveResource(
	ctx context.Context,
	resource domain.Resource,
	spec pluginx.ResourceSpec,
	path string,
) (pluginx.ResolvedResource, bool, error) {
	fields := resolveFields(resource, spec)
	missingFields := missingRequiredFields(spec.RequiredFields, fields)
	if len(missingFields) > 0 {
		return pluginx.ResolvedResource{}, false, newMissingInputError(missingFieldReasons(path, missingFields)...)
	}

	resolved := pluginx.ResolvedResource{
		ResourceID: resource.ID,
		ModelUID:   resource.ModelUID,
		Fields:     fields,
	}

	children, err := r.resolveSpecsWithPath(ctx, resource, spec.Children, false, path)
	if err != nil {
		return pluginx.ResolvedResource{}, false, err
	}
	if len(children) > 0 {
		resolved.Children = children
	}
	return resolved, true, nil
}

func emptyInput(spec pluginx.ResourceSpec) pluginx.ResolvedInput {
	return pluginx.ResolvedInput{
		Name:        spec.Name,
		Cardinality: spec.Cardinality,
		Resources:   []pluginx.ResolvedResource{},
	}
}

func (r *inputResolver) loadRelatedResources(ctx context.Context, base domain.Resource, spec pluginx.ResourceSpec) ([]domain.Resource, error) {
	relationName, err := buildRelationName(base.ModelUID, spec)
	if err != nil {
		return nil, err
	}

	ids, err := r.relatedIDs(ctx, base, spec, relationName)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}

	fields := specFields(spec)
	if len(spec.Filters) > 0 && spec.ModelUID != "" {
		resources, _, err := r.resources.ListResourcesWithFilters(
			ctx,
			fields,
			spec.ModelUID,
			ids,
			0,
			int64(len(ids)),
			filterGroups(spec.Filters),
		)
		return resources, err
	}

	resources, err := r.resources.ListResourceByIds(ctx, fields, ids)
	if err != nil {
		return nil, err
	}
	return filterResources(resources, spec), nil
}

func (r *inputResolver) relatedIDs(
	ctx context.Context,
	base domain.Resource,
	spec pluginx.ResourceSpec,
	relationName string,
) ([]int64, error) {
	switch spec.Direction {
	case pluginx.DirectionToTarget:
		return r.relations.ListSrcRelated(ctx, base.ModelUID, relationName, base.ID)
	case pluginx.DirectionToSource:
		return r.relations.ListDstRelated(ctx, base.ModelUID, relationName, base.ID)
	default:
		return nil, fmt.Errorf("插件关联方向不能为空: %s.%s", base.ModelUID, spec.Name)
	}
}

func resourceHasFields(resource domain.Resource, fields []string) bool {
	return lo.EveryBy(fields, func(field string) bool {
		_, ok := resource.Data[field]
		return ok
	})
}


func buildRelationName(baseModelUID string, spec pluginx.ResourceSpec) (string, error) {
	if spec.RelationType == "" {
		return "", fmt.Errorf("插件关联类型不能为空: %s.%s", baseModelUID, spec.Name)
	}
	if !pluginx.ValidRelationType(spec.RelationType) {
		return "", fmt.Errorf("插件关联类型不支持: %s", spec.RelationType)
	}
	if spec.ModelUID == "" {
		return "", fmt.Errorf("插件关联模型不能为空: %s.%s", baseModelUID, spec.Name)
	}

	switch spec.Direction {
	case pluginx.DirectionToTarget:
		return fmt.Sprintf("%s_%s_%s", baseModelUID, spec.RelationType, spec.ModelUID), nil
	case pluginx.DirectionToSource:
		return fmt.Sprintf("%s_%s_%s", spec.ModelUID, spec.RelationType, baseModelUID), nil
	default:
		return "", fmt.Errorf("插件关联方向不支持: %s", spec.Direction)
	}
}

func specFields(spec pluginx.ResourceSpec) []string {
	filterFields := lo.Map(spec.Filters, func(filter pluginx.Filter, _ int) string {
		return filter.Field
	})
	allFields := append(lo.Values(spec.Fields), filterFields...)
	return lo.Uniq(lo.Filter(allFields, func(field string, _ int) bool {
		return field != ""
	}))
}

func resolveFields(resource domain.Resource, spec pluginx.ResourceSpec) map[string]any {
	return lo.MapValues(spec.Fields, func(field string, _ string) any {
		return resource.Data[field]
	})
}

func missingRequiredFields(requirements []string, fields map[string]any) []string {
	return lo.Filter(requirements, func(fieldName string, _ int) bool {
		return !hasValue(fields[fieldName])
	})
}

func hasValue(value any) bool {
	switch v := value.(type) {
	case nil:
		return false
	case string:
		return v != ""
	default:
		return true
	}
}
