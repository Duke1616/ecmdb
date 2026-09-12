package codec

import (
	"reflect"
	"strings"

	"github.com/Duke1616/ecmdb/pkg/plugin/dsl"
	"github.com/Duke1616/ecmdb/pkg/plugin/types"
	"github.com/samber/lo"
)

// DeriveOption 用于向 DeriveSchema 传入模型级别的元数据覆写选项
type DeriveOption func(*deriveOptions)

type deriveOptions struct {
	// modelName  根模型的中文名称（若不设置则退回 modelUID）
	modelName string
	// modelGroup 根模型所属的模型分组名称（CMDB 分组，如 "安全认证"）
	modelGroup string
}

// WithModelName 设置根模型的中文展示名称
func WithModelName(name string) DeriveOption {
	return func(o *deriveOptions) {
		o.modelName = name
	}
}

// WithModelGroup 设置根模型所属的 CMDB 模型分组
func WithModelGroup(group string) DeriveOption {
	return func(o *deriveOptions) {
		o.modelGroup = group
	}
}

// DeriveSchema 基于泛型结构体 T 与入口 modelUID，自动反射推导出一套最小合法的 CMDB 元数据 Schema
// 自动推导模型、属性组、属性类型、安全字段标记（密码/私钥）以及模型间关联关系
func DeriveSchema[T any](modelUID string, opts ...DeriveOption) (types.Schema, error) {
	options := &deriveOptions{}
	for _, o := range opts {
		o(options)
	}

	spec, err := BuildCenterSpec[T]("target", modelUID)
	if err != nil {
		return types.Schema{}, err
	}

	var zero T
	rootType := peelType(reflect.TypeOf(zero))

	schema := types.Schema{
		RelationTypes: dsl.BasicRelationTypes(),
	}

	// 若声明了模型分组，先注入到 ModelGroups
	if options.modelGroup != "" {
		schema.ModelGroups = append(schema.ModelGroups, types.ModelGroupSpec{Name: options.modelGroup})
	}

	modelMap := make(map[string]*types.ModelSpec)
	collectModelsFromSpec(&schema, modelMap, spec, rootType)

	// 对根模型进行中文名称和分组的覆写
	if root, ok := modelMap[modelUID]; ok {
		if options.modelName != "" {
			root.Name = options.modelName
		}
		if options.modelGroup != "" {
			root.GroupName = options.modelGroup
		}
	}

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
			Name:    modelUID, // 默认 fallback 为 UID，可由 DeriveOption 覆写
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

		fieldUID := tag.field
		if fieldUID == "" {
			fieldUID = tag.name
		}

		attrType := inferAttributeType(field.Type)
		isSecure := isSecureField(fieldUID, tag.name)

		attrs = append(attrs, types.Attribute{
			UID:      fieldUID,
			Name:     tag.displayName(), // 优先使用 label 中文名，回退到 name（英文 UID）
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
