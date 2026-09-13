package codec

import (
	"fmt"
	"reflect"

	"github.com/Duke1616/ecmdb/pkg/plugin/types"
)

// BuildCenterSpec 基于中心节点模型结构 T 反射构建出根 ResourceSpec
func BuildCenterSpec[T any](name string, modelUID string) (types.ResourceSpec, error) {
	if name == "" {
		name = "target"
	}
	var zero T
	node, err := inspectStructNode(reflect.TypeOf(zero), name, modelUID, pluginTag{
		required: true,
	})
	if err != nil {
		return types.ResourceSpec{}, fmt.Errorf("BuildCenterSpec: %w", err)
	}
	return node.toResourceSpec(), nil
}
