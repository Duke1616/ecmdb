package plugin

import (
	"context"
	"errors"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/domain"
	pluginx "github.com/Duke1616/ecmdb/pkg/plugin"
)

// resourceReader 定义插件输入解析过程中需要读取的资源能力。
type resourceReader interface {
	// FindResourceById 查询单个资源，并只加载插件声明需要的字段。
	FindResourceById(ctx context.Context, fields []string, id int64) (domain.Resource, error)

	// ListResourceByIds 按资源 ID 批量查询关联资源，并只加载插件声明需要的字段。
	ListResourceByIds(ctx context.Context, fields []string, ids []int64) ([]domain.Resource, error)

	// ListResourcesWithFilters 按资源 ID 范围和过滤条件批量查询关联资源。
	ListResourcesWithFilters(ctx context.Context, fields []string, modelUID string, ids []int64, offset, limit int64, filterGroups []domain.FilterGroup) ([]domain.Resource, int64, error)
}

// relationReader 定义插件输入解析过程中需要读取的资源关联能力。
type relationReader interface {
	// ListSrcRelated 查询当前资源作为源端时，指定关联名称下的目标端资源 ID。
	ListSrcRelated(ctx context.Context, modelUID, relationName string, id int64) ([]int64, error)

	// ListDstRelated 查询当前资源作为目标端时，指定关联名称下的源端资源 ID。
	ListDstRelated(ctx context.Context, modelUID, relationName string, id int64) ([]int64, error)
}

type inputResolver struct {
	resources resourceReader
	relations relationReader
}

func newInputResolver(
	resources resourceReader,
	relations relationReader,
) *inputResolver {
	return &inputResolver{
		resources: resources,
		relations: relations,
	}
}

func (r *inputResolver) bindingSatisfied(ctx context.Context, primary domain.Resource, binding domain.PluginBinding) (bool, error) {
	_, err := r.resolve(ctx, primary, binding.Specs)
	if errors.Is(err, errRequiredInputMissing) {
		return false, nil
	}
	return err == nil, err
}

func (r *inputResolver) resolve(
	ctx context.Context,
	primary domain.Resource,
	specs []pluginx.ResourceSpec,
) (map[string]pluginx.ResolvedInput, error) {
	inputs, err := r.resolveSpecsWithPath(ctx, primary, specs, true, "")
	if err != nil {
		return nil, err
	}
	return inputs, nil
}

func (r *inputResolver) resolveSpecsWithPath(
	ctx context.Context,
	base domain.Resource,
	specs []pluginx.ResourceSpec,
	topLevel bool,
	parentPath string,
) (map[string]pluginx.ResolvedInput, error) {
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
		resource, err = r.loadResource(ctx, base.ID, fields)
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

func (r *inputResolver) loadResource(ctx context.Context, resourceID int64, fields []string) (domain.Resource, error) {
	if fields == nil {
		fields = []string{}
	}
	return r.resources.FindResourceById(ctx, fields, resourceID)
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
	for _, field := range fields {
		if _, ok := resource.Data[field]; !ok {
			return false
		}
	}
	return true
}
