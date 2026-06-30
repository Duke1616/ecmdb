package plugin

import (
	"fmt"
	"reflect"
	"strings"
)

// BuildSpecs 基于顶层输入容器 T 反射解析出 plugin.ResourceSpec 列表。
func BuildSpecs[T any]() ([]ResourceSpec, error) {
	var zero T
	t := peelType(reflect.TypeOf(zero))
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("BuildSpecs: T must be a struct, got %s", t.Kind())
	}

	var specs []ResourceSpec
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

func BuildCenterSpec[T any](name string, modelUID string) (ResourceSpec, error) {
	var zero T
	t := peelType(reflect.TypeOf(zero))
	if t.Kind() != reflect.Struct {
		return ResourceSpec{}, fmt.Errorf("BuildCenterSpec: T must be a struct, got %s", t.Kind())
	}
	if name == "" {
		name = "target"
	}

	spec := ResourceSpec{
		Name:        name,
		ModelUID:    modelUID,
		Cardinality: CardinalityOne,
		Required:    true,
		Fields:      make(map[string]string),
	}
	if err := fillSpecFieldsAndChildren(&spec, t); err != nil {
		return ResourceSpec{}, err
	}
	return spec, nil
}

// parseSpecFromField 将单个 struct 字段反射解析为 ResourceSpec 节点。
func parseSpecFromField(field reflect.StructField, tag pluginTag) (ResourceSpec, error) {
	cardinality := tag.cardinality
	if cardinality == "" {
		if peelType(field.Type).Kind() == reflect.Slice {
			cardinality = CardinalityMany
		} else {
			cardinality = CardinalityOne
		}
	}

	spec := ResourceSpec{
		Name:         tag.name,
		ModelUID:     tag.model,
		RelationType: tag.relationType,
		Direction:    tag.direction,
		Cardinality:  cardinality,
		Required:     tag.required,
		Fields:       make(map[string]string),
	}
	if spec.RelationType != "" && !ValidRelationType(spec.RelationType) {
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

// fillSpecFieldsAndChildren 递归地填充 ResourceSpec 节点的普通属性字段和嵌套关联的子节点。
func fillSpecFieldsAndChildren(spec *ResourceSpec, t reflect.Type) error {
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
			// 子关联节点：递归解析生成子 ResourceSpec
			childSpec, err := parseSpecFromField(field, tag)
			if err != nil {
				return err
			}
			spec.Children = append(spec.Children, childSpec)
		} else {
			// 普通字段属性
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

// Configurator 用于链式配置和修饰 ResourceSpec 节点树
type Configurator struct {
	specs []ResourceSpec
}

// Configure 初始化配置器
func Configure(specs []ResourceSpec) *Configurator {
	return &Configurator{specs: specs}
}

// ForPath 通过点路径（例如 "target.gateways.sub_gateways"）定位具体的 ResourceSpec 节点进行微调
func (c *Configurator) ForPath(path string, fn func(spec *ResourceSpec)) *Configurator {
	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return c
	}

	findAndModifySpec(c.specs, parts, fn)
	return c
}

// Build 返回微调配置后的 specs 列表
func (c *Configurator) Build() []ResourceSpec {
	return c.specs
}

func findAndModifySpec(specs []ResourceSpec, pathParts []string, fn func(spec *ResourceSpec)) bool {
	if len(pathParts) == 0 {
		return false
	}
	current := pathParts[0]
	for i := range specs {
		if specs[i].Name == current {
			if len(pathParts) == 1 {
				fn(&specs[i])
				return true
			}
			return findAndModifySpec(specs[i].Children, pathParts[1:], fn)
		}
	}
	return false
}
