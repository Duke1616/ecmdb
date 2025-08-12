package protox

import (
	"fmt"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
)

// NewAnyMap 将 map[string]interface{} 转成 map[string]*anypb.Any
func NewAnyMap(m map[string]interface{}) (map[string]*anypb.Any, error) {
	result := make(map[string]*anypb.Any, len(m))
	for k, v := range m {
		// 先转成 structpb.Value
		val, err := structpb.NewValue(v)
		if err != nil {
			return nil, fmt.Errorf("key %q: %w", k, err)
		}

		// 再封装成 anypb.Any
		anyVal, err := anypb.New(val)
		if err != nil {
			return nil, fmt.Errorf("key %q: %w", k, err)
		}

		result[k] = anyVal
	}
	return result, nil
}

func AnyMapToInterfaceMap(m map[string]*anypb.Any) (map[string]interface{}, error) {
	// 先把 map 转成 JSON（protojson 会帮你解 anypb.Any）
	b, err := protojson.Marshal(&structpb.Struct{Fields: anyMapToStructFields(m)})
	if err != nil {
		return nil, err
	}

	// 再把 JSON 转回 structpb.Struct
	var s structpb.Struct
	if err = protojson.Unmarshal(b, &s); err != nil {
		return nil, err
	}

	return s.AsMap(), nil
}

// 辅助函数，把 map[string]*anypb.Any 变成 map[string]*structpb.Value
func anyMapToStructFields(m map[string]*anypb.Any) map[string]*structpb.Value {
	res := make(map[string]*structpb.Value, len(m))
	for k, v := range m {
		if v == nil {
			res[k] = structpb.NewNullValue()
			continue
		}
		// 用 Any → JSON → Value
		jsonBytes, _ := protojson.Marshal(v)
		var val structpb.Value
		_ = protojson.Unmarshal(jsonBytes, &val)
		res[k] = &val
	}
	return res
}
