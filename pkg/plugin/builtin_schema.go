package plugin

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
)

func buildImportSchema(schema Schema, bindings []Binding) (Schema, error) {
	refs, err := collectSchemaRefs(bindings)
	if err != nil {
		return Schema{}, err
	}

	models := lo.Filter(schema.Models, func(model ModelSpec, _ int) bool {
		_, ok := refs.modelUIDs[strings.TrimSpace(model.UID)]
		return ok
	})

	groupNames := lo.SliceToMap(models, func(model ModelSpec) (string, struct{}) {
		return strings.TrimSpace(model.GroupName), struct{}{}
	})

	return Schema{
		ModelGroups: lo.Filter(schema.ModelGroups, func(group ModelGroupSpec, _ int) bool {
			_, ok := groupNames[strings.TrimSpace(group.Name)]
			return ok
		}),
		Models: models,
		RelationTypes: lo.Filter(schema.RelationTypes, func(relationType RelationType, _ int) bool {
			_, ok := refs.relationTypeUIDs[strings.TrimSpace(relationType.UID)]
			return ok
		}),
		ModelRelations: lo.Values(refs.relations),
	}, nil
}

type schemaRefs struct {
	modelUIDs        map[string]struct{}
	relationTypeUIDs map[string]struct{}
	relations        map[string]ModelRelation
}

func collectSchemaRefs(bindings []Binding) (schemaRefs, error) {
	refs := schemaRefs{
		modelUIDs:        make(map[string]struct{}),
		relationTypeUIDs: make(map[string]struct{}),
		relations:        make(map[string]ModelRelation),
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

func (refs schemaRefs) collectSpec(spec ResourceSpec) error {
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

func buildModelRelation(parent ResourceSpec, child ResourceSpec) (ModelRelation, bool, error) {
	relationType := strings.TrimSpace(child.RelationType)
	if relationType == "" {
		return ModelRelation{}, false, nil
	}

	sourceModelUID, targetModelUID, sourceCardinality, targetCardinality, err := relationEndpoints(parent, child)
	if err != nil {
		return ModelRelation{}, false, err
	}

	return ModelRelation{
		SourceModelUID:  sourceModelUID,
		TargetModelUID:  targetModelUID,
		RelationTypeUID: relationType,
		Mapping:         inferRelationMapping(sourceCardinality, targetCardinality),
	}, true, nil
}

func relationEndpoints(parent ResourceSpec, child ResourceSpec) (string, string, string, string, error) {
	switch child.Direction {
	case DirectionToSource:
		return child.ModelUID, parent.ModelUID, child.Cardinality, parent.Cardinality, nil
	case DirectionToTarget:
		return parent.ModelUID, child.ModelUID, parent.Cardinality, child.Cardinality, nil
	default:
		return "", "", "", "", fmt.Errorf("插件关联方向不支持: %s", child.Direction)
	}
}

func inferRelationMapping(sourceCardinality string, targetCardinality string) string {
	sourceCardinality = normalizeCardinality(sourceCardinality)
	targetCardinality = normalizeCardinality(targetCardinality)

	switch {
	case sourceCardinality == CardinalityOne && targetCardinality == CardinalityOne:
		return MappingOneToOne
	case sourceCardinality == CardinalityOne && targetCardinality == CardinalityMany:
		return MappingOneToMany
	default:
		return MappingManyToMany
	}
}

func normalizeCardinality(value string) string {
	if strings.TrimSpace(value) == "" {
		return CardinalityOne
	}
	return value
}
