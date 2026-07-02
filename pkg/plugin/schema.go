package plugin

type SetupItem interface {
	applyToSchema(schema *Schema)
}

type SchemaBuilder struct {
	schema Schema
}

func (b *SchemaBuilder) Setup(items ...SetupItem) *SchemaBuilder {
	for _, item := range items {
		if item != nil {
			item.applyToSchema(&b.schema)
		}
	}
	return b
}

func (b *SchemaBuilder) ModelGroup(name string) *SchemaBuilder {
	b.schema.ModelGroups = append(b.schema.ModelGroups, ModelGroupSpec{Name: name})
	return b
}

func (b *SchemaBuilder) ModelGroups(groups ...ModelGroupSpec) *SchemaBuilder {
	b.schema.ModelGroups = append(b.schema.ModelGroups, groups...)
	return b
}

func (b *SchemaBuilder) Model(model ModelSpec) *SchemaBuilder {
	b.schema.Models = append(b.schema.Models, model)
	return b
}

func (b *SchemaBuilder) Models(models ...ModelSpec) *SchemaBuilder {
	b.schema.Models = append(b.schema.Models, models...)
	return b
}

func (b *SchemaBuilder) RelationType(relationType RelationType) *SchemaBuilder {
	b.schema.RelationTypes = append(b.schema.RelationTypes, relationType)
	return b
}

func (b *SchemaBuilder) RelationTypes(relationTypes ...RelationType) *SchemaBuilder {
	b.schema.RelationTypes = append(b.schema.RelationTypes, relationTypes...)
	return b
}

func (b *SchemaBuilder) Relation(sourceModelUID, relationTypeUID, targetModelUID, mapping string) *SchemaBuilder {
	b.schema.ModelRelations = append(b.schema.ModelRelations, ModelRelation{
		SourceModelUID:  sourceModelUID,
		TargetModelUID:  targetModelUID,
		RelationTypeUID: relationTypeUID,
		Mapping:         mapping,
	})
	return b
}

func (b *SchemaBuilder) Build() Schema {
	return b.schema
}

type modelGroupItem string

func ModelGroup(name string) SetupItem {
	return modelGroupItem(name)
}

func (g modelGroupItem) applyToSchema(schema *Schema) {
	schema.ModelGroups = append(schema.ModelGroups, ModelGroupSpec{Name: string(g)})
}

type relationTypesItem []RelationType

func RelationTypes(relationTypes ...RelationType) SetupItem {
	return relationTypesItem(relationTypes)
}

func (r relationTypesItem) applyToSchema(schema *Schema) {
	schema.RelationTypes = append(schema.RelationTypes, r...)
}

type relationItem struct {
	relation ModelRelation
}

func Relation(sourceModelUID, relationTypeUID, targetModelUID string) *relationItem {
	return &relationItem{
		relation: ModelRelation{
			SourceModelUID:  sourceModelUID,
			TargetModelUID:  targetModelUID,
			RelationTypeUID: relationTypeUID,
			Mapping:         MappingManyToMany,
		},
	}
}

func (r *relationItem) OneToOne() *relationItem {
	r.relation.Mapping = MappingOneToOne
	return r
}

func (r *relationItem) OneToMany() *relationItem {
	r.relation.Mapping = MappingOneToMany
	return r
}

func (r *relationItem) ManyToMany() *relationItem {
	r.relation.Mapping = MappingManyToMany
	return r
}

func (r *relationItem) applyToSchema(schema *Schema) {
	schema.ModelRelations = append(schema.ModelRelations, r.relation)
}

func BasicRelationTypes() []RelationType {
	return []RelationType{
		{
			UID:            RelationTypeDefault,
			Name:           "默认关联",
			SourceDescribe: "关联",
			TargetDescribe: "关联",
		},
		{
			UID:            RelationTypeRun,
			Name:           "运行",
			SourceDescribe: "运行于",
			TargetDescribe: "运行",
		},
		{
			UID:            RelationTypeGroup,
			Name:           "组成",
			SourceDescribe: "组成",
			TargetDescribe: "组成于",
		},
		{
			UID:            RelationTypeBelong,
			Name:           "属于",
			SourceDescribe: "属于",
			TargetDescribe: "包含",
		},
	}
}
