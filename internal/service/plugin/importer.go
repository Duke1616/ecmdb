package plugin

import (
	"context"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/domain"
	attribute "github.com/Duke1616/ecmdb/internal/service/attribute"
	model "github.com/Duke1616/ecmdb/internal/service/model"
	relation "github.com/Duke1616/ecmdb/internal/service/relation"
	pluginx "github.com/Duke1616/ecmdb/pkg/plugin/types"
	"github.com/samber/lo"
)

// ISchemaImporter 负责将插件自描述中的动态 Schema（模型、属性、关系）幂等导入并同步到 CMDB 底层元数据库中
//
//go:generate mockgen -source=./importer.go -destination=./mocks/importer.mock.go -package=pluginmocks -typed ISchemaImporter
type ISchemaImporter interface {
	// ImportSchema 幂等编排并导入完整的插件自描述元数据（模型分组、模型、属性分组、字段定义、关系类型与模型间关联）
	ImportSchema(ctx context.Context, schema pluginx.Schema) error
}

type schemaImporter struct {
	models         model.Service
	modelGroups    model.MGService
	attributes     attribute.Service
	relationTypes  relation.RelationTypeService
	modelRelations relation.RelationModelService
}

func NewSchemaImporter(
	models model.Service,
	modelGroups model.MGService,
	attributes attribute.Service,
	relationTypes relation.RelationTypeService,
	modelRelations relation.RelationModelService,
) ISchemaImporter {
	return &schemaImporter{
		models:         models,
		modelGroups:    modelGroups,
		attributes:     attributes,
		relationTypes:  relationTypes,
		modelRelations: modelRelations,
	}
}

// ImportSchema 幂等编排导入完整的插件自描述元数据
func (i *schemaImporter) ImportSchema(ctx context.Context, schema pluginx.Schema) error {
	if err := i.importModels(ctx, schema.ModelGroups, schema.Models); err != nil {
		return err
	}
	if err := i.importRelationTypes(ctx, schema.RelationTypes); err != nil {
		return err
	}
	return i.importModelRelations(ctx, schema.ModelRelations)
}

func (i *schemaImporter) importModels(ctx context.Context, modelGroups []pluginx.ModelGroupSpec, models []pluginx.ModelSpec) error {
	groups, err := i.ensureModelGroups(ctx, modelGroups, models)
	if err != nil {
		return err
	}

	if err = i.ensureModels(ctx, models, groups); err != nil {
		return err
	}

	for _, modelSpec := range models {
		if err = i.ensureAttributes(ctx, modelSpec); err != nil {
			return err
		}
	}
	return nil
}

func (i *schemaImporter) ensureModelGroups(
	ctx context.Context,
	modelGroups []pluginx.ModelGroupSpec,
	models []pluginx.ModelSpec,
) (map[string]int64, error) {
	groupNames := append(
		lo.Map(modelGroups, func(g pluginx.ModelGroupSpec, _ int) string { return g.Name }),
		lo.Map(models, func(m pluginx.ModelSpec, _ int) string { return m.GroupName })...,
	)
	names := lo.Uniq(lo.Filter(groupNames, func(name string, _ int) bool {
		return name != ""
	}))
	if len(names) == 0 {
		return map[string]int64{}, nil
	}

	existing, err := i.modelGroups.GetByNames(ctx, names)
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
		created, err := i.modelGroups.BatchCreate(ctx, missing)
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

func (i *schemaImporter) ensureModels(
	ctx context.Context,
	models []pluginx.ModelSpec,
	groups map[string]int64,
) error {
	if len(models) == 0 {
		return nil
	}

	uids := lo.Map(models, func(m pluginx.ModelSpec, _ int) string { return m.UID })
	existingList, err := i.models.GetByUids(ctx, uids)
	if err != nil {
		return err
	}
	existingMap := lo.SliceToMap(existingList, func(m domain.Model) (string, domain.Model) {
		return m.UID, m
	})

	for _, spec := range models {
		if spec.UID == "" {
			return fmt.Errorf("插件模型 UID 不能为空")
		}
		if spec.Name == "" {
			return fmt.Errorf("插件模型名称不能为空: %s", spec.UID)
		}

		if _, exists := existingMap[spec.UID]; exists {
			// 已存在：跳过（模型名称变更需在 CMDB 控制台手动修改，或由未来迭代补全 Update 能力）
			continue
		}

		_, err = i.models.Create(ctx, domain.Model{
			UID:     spec.UID,
			Name:    spec.Name,
			Icon:    spec.Icon,
			GroupId: groups[spec.GroupName],
			Builtin: spec.Builtin,
		})
		if err != nil {
			return fmt.Errorf("创建插件模型失败 %s: %w", spec.UID, err)
		}
	}
	return nil
}

func (i *schemaImporter) ensureAttributes(ctx context.Context, model pluginx.ModelSpec) error {
	if len(model.AttributeGroups) == 0 {
		return nil
	}

	groups, err := i.ensureAttributeGroups(ctx, model)
	if err != nil {
		return err
	}
	return i.ensureAttributeFields(ctx, model, groups)
}

func (i *schemaImporter) ensureAttributeGroups(ctx context.Context, model pluginx.ModelSpec) (map[string]int64, error) {
	existing, err := i.attributes.ListAttributeGroup(ctx, model.UID)
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
		created, err := i.attributes.BatchCreateAttributeGroup(ctx, missing)
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

// fieldEntry 描述一个带 groupID 上下文的扁平化字段条目，供 ensureAttributeFields 的 upsert 流程消费
type fieldEntry struct {
	field   pluginx.Attribute
	groupID int64
}

func (i *schemaImporter) ensureAttributeFields(ctx context.Context, model pluginx.ModelSpec, groups map[string]int64) error {
	// 1. 前置校验并扁平化所有字段（保留 groupID 上下文）
	allFields, err := flattenAndValidateFields(model, groups)
	if err != nil {
		return err
	}

	// 2. 拉取已有属性并建立 UID → domain.Attribute 索引
	existing, _, err := i.attributes.ListAttributes(ctx, model.UID)
	if err != nil {
		return err
	}
	existingByUID := lo.SliceToMap(existing, func(attr domain.Attribute) (string, domain.Attribute) {
		return attr.FieldUid, attr
	})

	// 3. 更新名称已变化的 builtin 属性（因含 IO 副作用，采用 for range 而非 FlatMap）
	for _, attr := range lo.FilterMap(allFields, func(e fieldEntry, _ int) (domain.Attribute, bool) {
		old, exists := existingByUID[e.field.UID]
		if !exists || old.FieldName == e.field.Name {
			return domain.Attribute{}, false
		}
		updated := old
		updated.FieldName = e.field.Name
		return updated, true
	}) {
		if _, err = i.attributes.UpdateAttribute(ctx, attr); err != nil {
			return fmt.Errorf("更新插件属性名称失败 %s.%s: %w", model.UID, attr.FieldUid, err)
		}
	}

	// 4. 批量创建尚不存在的新字段
	toCreate := lo.FilterMap(allFields, func(e fieldEntry, _ int) (domain.Attribute, bool) {
		_, exists := existingByUID[e.field.UID]
		return domain.Attribute{
			GroupId:   e.groupID,
			ModelUid:  model.UID,
			FieldUid:  e.field.UID,
			FieldName: e.field.Name,
			FieldType: e.field.Type,
			Required:  e.field.Required,
			Display:   e.field.Display,
			Secure:    e.field.Secure,
			Index:     e.field.Index,
			SortKey:   e.field.Index,
			Option:    e.field.Option,
			Builtin:   e.field.Builtin,
		}, !exists
	})
	if len(toCreate) == 0 {
		return nil
	}
	return i.attributes.BatchCreateAttribute(ctx, toCreate)
}

// flattenAndValidateFields 前置校验字段 UID 非空，并将 AttributeGroups 扁平化为带 groupID 的 fieldEntry 切片
func flattenAndValidateFields(model pluginx.ModelSpec, groups map[string]int64) ([]fieldEntry, error) {
	var result []fieldEntry
	for _, group := range model.AttributeGroups {
		for _, field := range group.Fields {
			if field.UID == "" {
				return nil, fmt.Errorf("插件模型字段 UID 不能为空: %s.%s", model.UID, group.Name)
			}
			result = append(result, fieldEntry{field: field, groupID: groups[group.Name]})
		}
	}
	return result, nil
}

func (i *schemaImporter) importRelationTypes(ctx context.Context, relationTypes []pluginx.RelationType) error {
	if len(relationTypes) == 0 {
		return nil
	}

	uids := lo.Map(relationTypes, func(relationType pluginx.RelationType, _ int) string {
		return relationType.UID
	})
	existing, err := i.relationTypes.GetByUids(ctx, uids)
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
	return i.relationTypes.BatchCreate(ctx, missing)
}

func (i *schemaImporter) importModelRelations(ctx context.Context, relations []pluginx.ModelRelation) error {
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

	for idx := range modelRelations {
		if err := modelRelations[idx].Validate(); err != nil {
			return err
		}
	}

	names := lo.Map(modelRelations, func(relation domain.ModelRelation, _ int) string {
		return relation.RelationName
	})
	existing, err := i.modelRelations.GetByRelationNames(ctx, names)
	if err != nil {
		return err
	}
	existingByName := lo.SliceToMap(existing, func(relation domain.ModelRelation) (string, domain.ModelRelation) {
		return relation.RelationName, relation
	})

	toCreate := make([]domain.ModelRelation, 0, len(modelRelations))
	for _, relation := range modelRelations {
		existingRelation, ok := existingByName[relation.RelationName]
		if !ok {
			toCreate = append(toCreate, relation)
			continue
		}
		if modelRelationChanged(existingRelation, relation) {
			relation.ID = existingRelation.ID
			if _, err = i.modelRelations.UpdateModelRelation(ctx, relation); err != nil {
				return err
			}
		}
	}
	return i.modelRelations.BatchCreate(ctx, toCreate)
}

func modelRelationChanged(current domain.ModelRelation, next domain.ModelRelation) bool {
	return current.SourceModelUID != next.SourceModelUID ||
		current.TargetModelUID != next.TargetModelUID ||
		current.RelationTypeUID != next.RelationTypeUID ||
		current.Mapping != next.Mapping
}

func isEmptySchema(schema pluginx.Schema) bool {
	return len(schema.Models) == 0 &&
		len(schema.ModelGroups) == 0 &&
		len(schema.RelationTypes) == 0 &&
		len(schema.ModelRelations) == 0
}
