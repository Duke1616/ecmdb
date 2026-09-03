package codec

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/Duke1616/ecmdb/pkg/plugin/types"
	"github.com/samber/lo"
)

// InputOne 从动作上下文中读取单个输入，并解码成调用方声明的强类型结构 T
func InputOne[T any](actionCtx types.ActionContext, name string) (T, error) {
	var zero T
	input, ok := actionCtx.Inputs[name]
	if !ok || len(input.Resources) == 0 {
		return zero, fmt.Errorf("plugin input %s not found", name)
	}
	return DecodeResource[T](input.Resources[0])
}

// InputMany 从动作上下文中读取多个输入，并逐个解码成调用方声明的强类型结构 T
func InputMany[T any](actionCtx types.ActionContext, name string) ([]T, error) {
	input, ok := actionCtx.Inputs[name]
	if !ok {
		return nil, fmt.Errorf("plugin input %s not found", name)
	}
	return lo.MapErr(input.Resources, func(r types.ResolvedResource, _ int) (T, error) {
		return DecodeResource[T](r)
	})
}

// RootInputName 自动推断根节点输入名称
func RootInputName(actionCtx types.ActionContext) (string, error) {
	if actionCtx.Binding.Graph != nil {
		for _, node := range actionCtx.Binding.Graph.Nodes {
			if node.ID == actionCtx.Binding.Graph.EntryNodeID && node.Name != "" {
				return node.Name, nil
			}
		}
	}
	if len(actionCtx.Inputs) == 1 {
		for name := range actionCtx.Inputs {
			return name, nil
		}
	}
	return "", fmt.Errorf("plugin root input is ambiguous")
}

// InputRootOne 自动定位入口输入并解码为目标结构体 T
func InputRootOne[T any](actionCtx types.ActionContext) (T, error) {
	name, err := RootInputName(actionCtx)
	if err != nil {
		var zero T
		return zero, err
	}
	return InputOne[T](actionCtx, name)
}

// DecodeResource 将已解析资源按照 `plugin` tag 解码成强类型结构 T
func DecodeResource[T any](resource types.ResolvedResource) (T, error) {
	var out T
	rv := reflect.ValueOf(&out)
	if err := decodeInto(rv, resource); err != nil {
		return out, err
	}
	return out, nil
}

// decodeInto 递归解析并为目标 dst 赋值
func decodeInto(dst reflect.Value, resource types.ResolvedResource) error {
	dst = peelValue(dst)

	if dst.Kind() != reflect.Struct {
		return fmt.Errorf("plugin decode target must be struct, got %s", dst.Kind())
	}

	rt := dst.Type()
	for i := 0; i < rt.NumField(); i++ {
		fieldType := rt.Field(i)
		field := dst.Field(i)

		// 优先处理匿名嵌套结构体 (Embedded Struct) 的展开
		if fieldType.Anonymous && fieldType.Tag.Get("plugin") == "" {
			inner := peelValue(field)
			if inner.Kind() == reflect.Struct {
				if err := decodeInto(field, resource); err != nil {
					return err
				}
				continue
			}
		}

		// 忽略非匿名的未导出字段
		if fieldType.PkgPath != "" {
			continue
		}

		if err := decodeField(field, fieldType, resource); err != nil {
			return err
		}
	}
	return nil
}

// decodeField 解析单个字段
func decodeField(field reflect.Value, fieldType reflect.StructField, resource types.ResolvedResource) error {
	tag := parsePluginTag(fieldType)
	if tag.skip {
		return nil
	}

	// 1. 如果该字段被标识为子级关联资源，则按子级解析
	if child, ok := resource.Children[tag.name]; ok {
		if err := setChildValue(field, child); err != nil {
			return fmt.Errorf("decode child %s: %w", tag.name, err)
		}
		return nil
	}

	// 2. 普通属性字段解析
	val, ok := resource.Fields[tag.name]
	if !ok || val == nil {
		if tag.defaultValue != "" {
			val = tag.defaultValue
		} else if tag.required {
			return fmt.Errorf("plugin field %s is required", tag.name)
		} else {
			return nil
		}
	}

	if err := setFieldValue(field, val); err != nil {
		return fmt.Errorf("decode field %s: %w", tag.name, err)
	}

	return nil
}

// setChildValue 为嵌套的子资源赋值（可以是 Slice 也可以是 Struct）
func setChildValue(dst reflect.Value, input types.ResolvedInput) error {
	dst = peelValue(dst)

	switch dst.Kind() {
	case reflect.Slice:
		slice := reflect.MakeSlice(dst.Type(), 0, len(input.Resources))
		for _, resource := range input.Resources {
			item := reflect.New(dst.Type().Elem()).Elem()
			if err := decodeInto(item, resource); err != nil {
				return err
			}
			slice = reflect.Append(slice, item)
		}
		dst.Set(slice)
		return nil
	case reflect.Struct:
		if len(input.Resources) == 0 {
			return nil
		}
		return decodeInto(dst, input.Resources[0])
	default:
		return fmt.Errorf("child target must be struct or slice, got %s", dst.Kind())
	}
}

// setFieldValue 实现弱类型的安全赋值转换
func setFieldValue(dst reflect.Value, value any) error {
	dst = peelValue(dst)
	if !dst.CanSet() {
		return nil
	}

	src := reflect.ValueOf(value)
	if src.IsValid() {
		if src.Type().AssignableTo(dst.Type()) {
			dst.Set(src)
			return nil
		}
		if src.Type().ConvertibleTo(dst.Type()) {
			dst.Set(src.Convert(dst.Type()))
			return nil
		}
	}

	strVal := fmt.Sprint(value)
	switch dst.Kind() {
	case reflect.String:
		dst.SetString(strVal)
	case reflect.Bool:
		v, err := strconv.ParseBool(strVal)
		if err != nil {
			return err
		}
		dst.SetBool(v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(strVal, 10, dst.Type().Bits())
		if err != nil {
			return err
		}
		dst.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(strVal, 10, dst.Type().Bits())
		if err != nil {
			return err
		}
		dst.SetUint(v)
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(strVal, dst.Type().Bits())
		if err != nil {
			return err
		}
		dst.SetFloat(v)
	default:
		return fmt.Errorf("unsupported target kind %s", dst.Kind())
	}
	return nil
}

// ── 反射工具函数 ─────────────────────────────────────────────────────────────

func peelType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func peelValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() && v.CanSet() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v
}

func isStructOrStructSlice(t reflect.Type) (bool, reflect.Type) {
	unwrapped := peelType(t)
	switch unwrapped.Kind() {
	case reflect.Struct:
		return true, unwrapped
	case reflect.Slice, reflect.Array:
		elemType := peelType(unwrapped.Elem())
		if elemType.Kind() == reflect.Struct {
			return true, elemType
		}
	}
	return false, nil
}
