package plugin

import (
	"github.com/Duke1616/ecmdb/pkg/plugin/codec"
	"github.com/Duke1616/ecmdb/pkg/plugin/dsl"
	"github.com/Duke1616/ecmdb/pkg/plugin/types"
	"github.com/samber/lo"
)

type derivedSchemaItem[T any] struct {
	modelUID string
}

func (d *derivedSchemaItem[T]) ApplyToSchema(schema *types.Schema) {
	derived, err := codec.DeriveSchema[T](d.modelUID)
	if err != nil {
		panic(err)
	}

	// 1. 合并模型列表（已有同 UID 模型不覆盖）
	existingModelUIDs := lo.SliceToMap(schema.Models, func(m types.ModelSpec) (string, struct{}) {
		return m.UID, struct{}{}
	})
	for _, model := range derived.Models {
		if _, exists := existingModelUIDs[model.UID]; !exists {
			schema.Models = append(schema.Models, model)
		}
	}

	// 2. 合并关联类型定义
	existingRelationUIDs := lo.SliceToMap(schema.RelationTypes, func(r types.RelationType) (string, struct{}) {
		return r.UID, struct{}{}
	})
	for _, relType := range derived.RelationTypes {
		if _, exists := existingRelationUIDs[relType.UID]; !exists {
			schema.RelationTypes = append(schema.RelationTypes, relType)
		}
	}

	// 3. 合并模型间关联关系
	for _, rel := range derived.ModelRelations {
		exists := lo.SomeBy(schema.ModelRelations, func(existing types.ModelRelation) bool {
			return existing.SourceModelUID == rel.SourceModelUID &&
				existing.TargetModelUID == rel.TargetModelUID &&
				existing.RelationTypeUID == rel.RelationTypeUID
		})
		if !exists {
			schema.ModelRelations = append(schema.ModelRelations, rel)
		}
	}
}

// Derive 基于目标结构体 T 及其根模型 UID，自动反射推导并注入所需的 CMDB 模型、属性和关联关系
// 一行代码直接替代繁冗的 Setup(...) 手写 Schema 定义，实现零重复的约定优于配置
func Derive[T any](modelUID string) dsl.SetupItem {
	return &derivedSchemaItem[T]{modelUID: modelUID}
}
