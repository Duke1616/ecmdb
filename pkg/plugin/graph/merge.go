package graph

import (
	"fmt"
	"strings"

	"github.com/Duke1616/ecmdb/pkg/plugin/types"
	"github.com/samber/lo"
)

// BuildImportSchema 根据当前配置的 bindings 引用，智能裁剪并组装需要导入 CMDB 的元数据 Schema
func BuildImportSchema(schema types.Schema, bindings []types.Binding) (types.Schema, error) {
	refs, err := collectSchemaRefs(bindings)
	if err != nil {
		return types.Schema{}, err
	}

	models := lo.Filter(schema.Models, func(model types.ModelSpec, _ int) bool {
		_, ok := refs.modelUIDs[strings.TrimSpace(model.UID)]
		return ok
	})

	groupNames := lo.SliceToMap(models, func(model types.ModelSpec) (string, struct{}) {
		return strings.TrimSpace(model.GroupName), struct{}{}
	})

	return types.Schema{
		ModelGroups: lo.Filter(schema.ModelGroups, func(group types.ModelGroupSpec, _ int) bool {
			_, ok := groupNames[strings.TrimSpace(group.Name)]
			return ok
		}),
		Models: models,
		RelationTypes: lo.Filter(schema.RelationTypes, func(relationType types.RelationType, _ int) bool {
			_, ok := refs.relationTypeUIDs[strings.TrimSpace(relationType.UID)]
			return ok
		}),
		ModelRelations: lo.Values(refs.relations),
	}, nil
}

type schemaRefs struct {
	modelUIDs        map[string]struct{}
	relationTypeUIDs map[string]struct{}
	relations        map[string]types.ModelRelation
}

func collectSchemaRefs(bindings []types.Binding) (schemaRefs, error) {
	refs := schemaRefs{
		modelUIDs:        make(map[string]struct{}),
		relationTypeUIDs: make(map[string]struct{}),
		relations:        make(map[string]types.ModelRelation),
	}

	for _, binding := range bindings {
		if binding.Graph == nil {
			continue
		}

		specs, err := CompileBindingGraph(binding.Graph)
		if err != nil {
			return schemaRefs{}, err
		}
		for _, spec := range specs {
			if err = refs.collectSpec(spec); err != nil {
				return schemaRefs{}, err
			}
		}
	}
	return refs, nil
}

func (refs schemaRefs) collectSpec(spec types.ResourceSpec) error {
	if modelUID := strings.TrimSpace(spec.ModelUID); modelUID != "" {
		refs.modelUIDs[modelUID] = struct{}{}
	}

	for _, child := range spec.Children {
		relation, ok, err := buildModelRelation(spec, child)
		if err != nil {
			return err
		}
		if ok {
			key := relation.SourceModelUID + "|" + relation.RelationTypeUID + "|" + relation.TargetModelUID + "|" + relation.Mapping
			refs.relations[key] = relation
			refs.relationTypeUIDs[relation.RelationTypeUID] = struct{}{}
		}

		if err = refs.collectSpec(child); err != nil {
			return err
		}
	}
	return nil
}

func buildModelRelation(parent types.ResourceSpec, child types.ResourceSpec) (types.ModelRelation, bool, error) {
	relationType := strings.TrimSpace(child.RelationType)
	if relationType == "" {
		return types.ModelRelation{}, false, nil
	}

	sourceModelUID, targetModelUID, sourceCardinality, targetCardinality, err := relationEndpoints(parent, child)
	if err != nil {
		return types.ModelRelation{}, false, err
	}

	return types.ModelRelation{
		SourceModelUID:  sourceModelUID,
		TargetModelUID:  targetModelUID,
		RelationTypeUID: relationType,
		Mapping:         inferRelationMapping(sourceCardinality, targetCardinality),
	}, true, nil
}

func relationEndpoints(parent types.ResourceSpec, child types.ResourceSpec) (string, string, string, string, error) {
	switch child.Direction {
	case types.DirectionToSource:
		return child.ModelUID, parent.ModelUID, child.Cardinality, parent.Cardinality, nil
	case types.DirectionToTarget:
		return parent.ModelUID, child.ModelUID, parent.Cardinality, child.Cardinality, nil
	default:
		return "", "", "", "", fmt.Errorf("插件关联方向不支持: %s", child.Direction)
	}
}

func inferRelationMapping(sourceCardinality string, targetCardinality string) string {
	sourceCardinality = normalizeCardinality(sourceCardinality)
	targetCardinality = normalizeCardinality(targetCardinality)

	switch {
	case sourceCardinality == types.CardinalityOne && targetCardinality == types.CardinalityOne:
		return types.MappingOneToOne
	case sourceCardinality == types.CardinalityOne && targetCardinality == types.CardinalityMany:
		return types.MappingOneToMany
	default:
		return types.MappingManyToMany
	}
}

func normalizeCardinality(value string) string {
	if strings.TrimSpace(value) == "" {
		return types.CardinalityOne
	}
	return value
}
