package plugin

import (
	"context"
	"errors"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/domain"
	"github.com/Duke1616/ecmdb/internal/errs"
	pluginx "github.com/Duke1616/ecmdb/pkg/plugin"
	"github.com/samber/lo"
)

func (s *service) importDefinition(ctx context.Context, def pluginx.Definition) error {
	if err := s.importSchema(ctx, def.Schema); err != nil {
		return err
	}
	return s.importPluginDefinition(ctx, def)
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
	existingNames := lo.SliceToMap(existing, func(relation domain.ModelRelation) (string, struct{}) {
		return relation.RelationName, struct{}{}
	})

	missing := lo.Filter(modelRelations, func(relation domain.ModelRelation, _ int) bool {
		_, ok := existingNames[relation.RelationName]
		return !ok
	})
	return s.modelRelations.BatchCreate(ctx, missing)
}

func (s *service) importPluginDefinition(ctx context.Context, def pluginx.Definition) error {
	if err := s.syncBuiltinPlugin(ctx, def.Plugin); err != nil {
		return err
	}

	for _, binding := range def.Bindings {
		if err := s.syncBuiltinBinding(ctx, binding); err != nil {
			return err
		}
	}
	return nil
}
