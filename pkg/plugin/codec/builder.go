package codec

import (
	"fmt"
	"reflect"

	"github.com/Duke1616/ecmdb/pkg/plugin/types"
)

// BuildSpecs 基于顶层输入容器 T 反射解析出 types.ResourceSpec 列表
func BuildSpecs[T any]() ([]types.ResourceSpec, error) {
	var zero T
	t := peelType(reflect.TypeOf(zero))
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("BuildSpecs: T must be a struct, got %s", t.Kind())
	}

	var specs []types.ResourceSpec
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" { // 忽略未导出字段
			continue
		}

		tag := parsePluginTag(field)
		if tag.skip {
			continue
		}

		spec, err := parseSpecFromField(field, tag)
		if err != nil {
			return nil, fmt.Errorf("parse field %s error: %w", field.Name, err)
		}
		specs = append(specs, spec)
	}

	return specs, nil
}

// BuildCenterSpec 基于中心节点模型结构 T 反射构建出根 ResourceSpec
func BuildCenterSpec[T any](name string, modelUID string) (types.ResourceSpec, error) {
	var zero T
	t := peelType(reflect.TypeOf(zero))
	if t.Kind() != reflect.Struct {
		return types.ResourceSpec{}, fmt.Errorf("BuildCenterSpec: T must be a struct, got %s", t.Kind())
	}
	if name == "" {
		name = "target"
	}

	spec := types.ResourceSpec{
		Name:        name,
		ModelUID:    modelUID,
		Cardinality: types.CardinalityOne,
		Required:    true,
		Fields:      make(map[string]string),
	}
	if err := fillSpecFieldsAndChildren(&spec, t); err != nil {
		return types.ResourceSpec{}, err
	}
	return spec, nil
}

// parseSpecFromField 将单个 struct 字段反射解析为 ResourceSpec 节点
func parseSpecFromField(field reflect.StructField, tag pluginTag) (types.ResourceSpec, error) {
	cardinality := tag.cardinality
	if cardinality == "" {
		if peelType(field.Type).Kind() == reflect.Slice {
			cardinality = types.CardinalityMany
		} else {
			cardinality = types.CardinalityOne
		}
	}

	spec := types.ResourceSpec{
		Name:         tag.name,
		ModelUID:     tag.model,
		RelationType: tag.relationType,
		Direction:    tag.direction,
		Cardinality:  cardinality,
		Required:     tag.required,
		Fields:       make(map[string]string),
	}
	if spec.RelationType != "" && !types.ValidRelationType(spec.RelationType) {
		return spec, fmt.Errorf("unsupported relation type %s", spec.RelationType)
	}

	// 递归解析结构体子节点
	underlyingType := peelType(field.Type)
	if underlyingType.Kind() == reflect.Slice {
		underlyingType = peelType(underlyingType.Elem())
	}

	if underlyingType.Kind() == reflect.Struct {
		if err := fillSpecFieldsAndChildren(&spec, underlyingType); err != nil {
			return spec, err
		}
	}

	return spec, nil
}

// fillSpecFieldsAndChildren 递归地填充 ResourceSpec 节点的普通属性字段和嵌套关联的子节点
func fillSpecFieldsAndChildren(spec *types.ResourceSpec, t reflect.Type) error {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}

		// 处理匿名嵌套结构体 (Embedded Struct) 的情况
		if field.Anonymous && field.Tag.Get("plugin") == "" {
			underlying := peelType(field.Type)
			if underlying.Kind() == reflect.Struct {
				if err := fillSpecFieldsAndChildren(spec, underlying); err != nil {
					return err
				}
			}
			continue
		}

		tag := parsePluginTag(field)
		if tag.skip {
			continue
		}

		// 判断是子关联节点还是普通属性字段
		if ok, _ := isStructOrStructSlice(field.Type); ok {
			childSpec, err := parseSpecFromField(field, tag)
			if err != nil {
				return err
			}
			spec.Children = append(spec.Children, childSpec)
		} else {
			cmdbField := tag.field
			if cmdbField == "" {
				cmdbField = tag.name
			}
			spec.Fields[tag.name] = cmdbField

			if tag.required {
				spec.RequiredFields = append(spec.RequiredFields, tag.name)
			}
		}
	}
	return nil
}
