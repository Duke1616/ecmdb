package protox

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

func AnyMapToInterfaceMap(m *structpb.Struct) (map[string]interface{}, error) {
	b, err := protojson.Marshal(m)
	if err != nil {
		return nil, err
	}

	var s structpb.Struct
	if err = protojson.Unmarshal(b, &s); err != nil {
		return nil, err
	}

	return s.AsMap(), nil
}

func ToMapInterface(labels map[string]string) map[string]interface{} {
	res := make(map[string]interface{}, len(labels))
	for k, v := range labels {
		res[k] = v
	}
	return res
}

func ToInterfaceSlice[T any](arr []T) []interface{} {
	res := make([]interface{}, len(arr))
	for i, v := range arr {
		res[i] = v
	}
	return res
}
