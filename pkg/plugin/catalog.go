package plugin

import (
	"github.com/Duke1616/ecmdb/pkg/plugin/graph"
	"github.com/Duke1616/ecmdb/pkg/plugin/types"
)

// Builtin 表示一个内置插件，提供完整 Definition 与按 Bindings 裁剪 Schema 的能力
type Builtin interface {
	Definition() Definition
	SchemaForBindings(bindings []types.Binding) (types.Schema, error)
}

type staticBuiltin struct {
	def Definition
}

// StaticBuiltin 将一个已构建好的 Definition 包装成 Builtin 实现
func StaticBuiltin(def Definition) Builtin {
	return staticBuiltin{def: def}
}

func (b staticBuiltin) Definition() Definition {
	return b.def
}

func (b staticBuiltin) SchemaForBindings(bindings []types.Binding) (types.Schema, error) {
	return graph.BuildImportSchema(b.def.Schema, bindings)
}
