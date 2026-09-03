package codec

import (
	"reflect"
	"strings"

	"github.com/Duke1616/ecmdb/pkg/plugin/dsl"
	"github.com/Duke1616/ecmdb/pkg/plugin/types"
	"github.com/samber/lo"
)

// DeriveSchema 基于泛型结构体 T 与入口 modelUID，自动反射推导出一套最小合法的 CMDB 元数据 Schema
// 自动推导模型、属性组、属性类型、安全字段标记（密码/私钥）以及模型间关联关系
func DeriveSchema[T any](modelUID string) (types.Schema, error) {
	spec, err := BuildCenterSpec[T]("target", modelUID)
	if err != nil {
		return types.Schema{}, err
	}

	var zero T
	rootType := peelType(reflect.TypeOf(zero))

	schema := types.Schema{
		RelationTypes: dsl.BasicRelationTypes(),
	}

	modelMap := make(map[string]*types.ModelSpec)
	collectModelsFromSpec(&schema, modelMap, spec, rootType)

	schema.Models = lo.Map(lo.Values(modelMap), func(m *types.ModelSpec, _ int) types.ModelSpec {
		return *m
	})

	return schema, nil
}

func collectModelsFromSpec(
	schema *types.Schema,
	modelMap map[string]*types.ModelSpec,
	spec types.ResourceSpec,
	t reflect.Type,
) {
	modelUID := strings.TrimSpace(spec.ModelUID)
	if modelUID == "" {
		return
	}

	if _, exists := modelMap[modelUID]; !exists {
		modelSpec := &types.ModelSpec{
			UID:     modelUID,
			Name:    modelUID,
			Builtin: true,
			AttributeGroups: []types.AttributeGroup{
				{
					Name:   "基础属性",
					Index:  0,
					Fields: extractAttributesFromType(t),
				},
			},
		}
		modelMap[modelUID] = modelSpec
	}

	// 遍历子关联，推导 ModelRelation 并递归子结构体
	for _, child := range spec.Children {
		childType := findFieldTypeByName(t, child.Name)
		if childType != nil {
			collectModelsFromSpec(schema, modelMap, child, childType)
		}

		if child.RelationType != "" && child.ModelUID != "" {
			relation := types.ModelRelation{
				RelationTypeUID: child.RelationType,
				Mapping:         types.MappingManyToMany,
			}
			if child.Direction == types.DirectionToSource {
				relation.SourceModelUID = child.ModelUID
				relation.TargetModelUID = modelUID
			} else {
				relation.SourceModelUID = modelUID
				relation.TargetModelUID = child.ModelUID
			}

			// 避免重复追加
			exists := lo.SomeBy(schema.ModelRelations, func(r types.ModelRelation) bool {
				return r.SourceModelUID == relation.SourceModelUID &&
					r.TargetModelUID == relation.TargetModelUID &&
					r.RelationTypeUID == relation.RelationTypeUID
			})
			if !exists {
				schema.ModelRelations = append(schema.ModelRelations, relation)
			}
		}
	}
}

func extractAttributesFromType(t reflect.Type) []types.Attribute {
	t = peelType(t)
	if t.Kind() != reflect.Struct {
		return nil
	}

	attrs := make([]types.Attribute, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}

		// 处理匿名内嵌结构体
		if field.Anonymous && field.Tag.Get("plugin") == "" {
			attrs = append(attrs, extractAttributesFromType(field.Type)...)
			continue
		}

		tag := parsePluginTag(field)
		if tag.skip {
			continue
		}

		// 忽略子模型关联切片/嵌套
		if ok, _ := isStructOrStructSlice(field.Type); ok {
			continue
		}

		fieldName := tag.field
		if fieldName == "" {
			fieldName = tag.name
		}

		attrType := inferAttributeType(field.Type)
		isSecure := isSecureField(fieldName, tag.name)

		attrs = append(attrs, types.Attribute{
			UID:      fieldName,
			Name:     fieldName,
			Type:     attrType,
			Required: tag.required,
			Display:  !isSecure,
			Secure:   isSecure,
			Builtin:  true,
			Index:    int64(len(attrs) + 1),
		})
	}
	return attrs
}

func inferAttributeType(t reflect.Type) string {
	t = peelType(t)
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "int"
	case reflect.Float32, reflect.Float64:
		return "float"
	case reflect.Bool:
		return "bool"
	default:
		return "string"
	}
}

func isSecureField(fieldName, tagName string) bool {
	lower := strings.ToLower(fieldName + " " + tagName)
	return strings.Contains(lower, "password") ||
		strings.Contains(lower, "private_key") ||
		strings.Contains(lower, "secret") ||
		strings.Contains(lower, "token")
}

func findFieldTypeByName(t reflect.Type, name string) reflect.Type {
	t = peelType(t)
	if t.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Anonymous {
			if found := findFieldTypeByName(field.Type, name); found != nil {
				return found
			}
		}
		tag := parsePluginTag(field)
		if tag.name == name {
			elem := peelType(field.Type)
			if elem.Kind() == reflect.Slice || elem.Kind() == reflect.Array {
				return peelType(elem.Elem())
			}
			return elem
		}
	}
	return nil
}
