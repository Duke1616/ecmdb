package plugin

import "reflect"

// peelType 解开类型中的所有指针封装，返回底层的非指针类型。
func peelType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

// peelValue 自动初始化并解开 Value 的所有指针封装，返回可直接赋值的底层 Value。
func peelValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() && v.CanSet() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v
}

// isStructOrStructSlice 判定指定类型是否为结构体或结构体切片/数组，并返回其底层的真实结构体 Type。
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
	default:
	}
	return false, nil
}
