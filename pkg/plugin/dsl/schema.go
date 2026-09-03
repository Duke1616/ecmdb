package dsl

import (
	"github.com/Duke1616/ecmdb/pkg/plugin/types"
)

// SetupItem 描述能够对 Schema 执行配置注入的项目接口
type SetupItem interface {
	ApplyToSchema(schema *types.Schema)
}

type SchemaBuilder struct {
	schema types.Schema
}

func (b *SchemaBuilder) Setup(items ...SetupItem) *SchemaBuilder {
	for _, item := range items {
		if item != nil {
			item.ApplyToSchema(&b.schema)
		}
	}
	return b
}

func (b *SchemaBuilder) ModelGroup(name string) *SchemaBuilder {
	b.schema.ModelGroups = append(b.schema.ModelGroups, types.ModelGroupSpec{Name: name})
	return b
}

func (b *SchemaBuilder) ModelGroups(groups ...types.ModelGroupSpec) *SchemaBuilder {
	b.schema.ModelGroups = append(b.schema.ModelGroups, groups...)
	return b
}

func (b *SchemaBuilder) Model(model types.ModelSpec) *SchemaBuilder {
	b.schema.Models = append(b.schema.Models, model)
	return b
}

func (b *SchemaBuilder) Models(models ...types.ModelSpec) *SchemaBuilder {
	b.schema.Models = append(b.schema.Models, models...)
	return b
}

func (b *SchemaBuilder) RelationType(relationType types.RelationType) *SchemaBuilder {
	b.schema.RelationTypes = append(b.schema.RelationTypes, relationType)
	return b
}

func (b *SchemaBuilder) RelationTypes(relationTypes ...types.RelationType) *SchemaBuilder {
	b.schema.RelationTypes = append(b.schema.RelationTypes, relationTypes...)
	return b
}

func (b *SchemaBuilder) Relation(sourceModelUID, relationTypeUID, targetModelUID, mapping string) *SchemaBuilder {
	b.schema.ModelRelations = append(b.schema.ModelRelations, types.ModelRelation{
		SourceModelUID:  sourceModelUID,
		TargetModelUID:  targetModelUID,
		RelationTypeUID: relationTypeUID,
		Mapping:         mapping,
	})
	return b
}

func (b *SchemaBuilder) Build() types.Schema {
	return b.schema
}

type modelGroupItem string

func ModelGroup(name string) SetupItem {
	return modelGroupItem(name)
}

func (g modelGroupItem) ApplyToSchema(schema *types.Schema) {
	schema.ModelGroups = append(schema.ModelGroups, types.ModelGroupSpec{Name: string(g)})
}

type relationTypesItem []types.RelationType

func RelationTypes(relationTypes ...types.RelationType) SetupItem {
	return relationTypesItem(relationTypes)
}

func (r relationTypesItem) ApplyToSchema(schema *types.Schema) {
	schema.RelationTypes = append(schema.RelationTypes, r...)
}

type relationItem struct {
	relation types.ModelRelation
}

func Relation(sourceModelUID, relationTypeUID, targetModelUID string) *relationItem {
	return &relationItem{
		relation: types.ModelRelation{
			SourceModelUID:  sourceModelUID,
			TargetModelUID:  targetModelUID,
			RelationTypeUID: relationTypeUID,
			Mapping:         types.MappingManyToMany,
		},
	}
}

func (r *relationItem) OneToOne() *relationItem {
	r.relation.Mapping = types.MappingOneToOne
	return r
}

func (r *relationItem) OneToMany() *relationItem {
	r.relation.Mapping = types.MappingOneToMany
	return r
}

func (r *relationItem) ManyToMany() *relationItem {
	r.relation.Mapping = types.MappingManyToMany
	return r
}

func (r *relationItem) ApplyToSchema(schema *types.Schema) {
	schema.ModelRelations = append(schema.ModelRelations, r.relation)
}

// BasicRelationTypes 返回 ECMDB 内置预设的基础关联类型定义
func BasicRelationTypes() []types.RelationType {
	return []types.RelationType{
		{
			UID:            types.RelationTypeDefault,
			Name:           "默认关联",
			SourceDescribe: "关联",
			TargetDescribe: "关联",
		},
		{
			UID:            types.RelationTypeRun,
			Name:           "运行",
			SourceDescribe: "运行于",
			TargetDescribe: "运行",
		},
		{
			UID:            types.RelationTypeGroup,
			Name:           "组成",
			SourceDescribe: "组成",
			TargetDescribe: "组成于",
		},
		{
			UID:            types.RelationTypeBelong,
			Name:           "属于",
			SourceDescribe: "属于",
			TargetDescribe: "包含",
		},
	}
}
