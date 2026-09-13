package codec

import (
	"github.com/Duke1616/ecmdb/pkg/plugin/types"
)

// DeriveSchema 基于泛型结构体 T 与入口 modelUID，自动推导最小合法的 CMDB 元数据 Schema
func DeriveSchema[T any](modelUID string) (types.Schema, error) {
	meta, err := InspectTarget[T]("target", modelUID)
	if err != nil {
		return types.Schema{}, err
	}
	return meta.Schema, nil
}
