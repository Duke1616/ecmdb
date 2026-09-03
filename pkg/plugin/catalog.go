package plugin

import (
	"fmt"
	"sort"
	"sync"

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

// ── 全局内置插件目录 ──────────────────────────────────────────────────────────

var builtinCatalog = struct {
	sync.RWMutex
	builtins map[string]Builtin
}{
	builtins: make(map[string]Builtin),
}

// RegisterBuiltin 注册内置插件
func RegisterBuiltin(builtin Builtin) error {
	if builtin == nil {
		return fmt.Errorf("builtin plugin is nil")
	}

	def := builtin.Definition()
	if def.Plugin.UID == "" {
		return fmt.Errorf("plugin uid is required")
	}

	builtinCatalog.Lock()
	defer builtinCatalog.Unlock()

	if _, exists := builtinCatalog.builtins[def.Plugin.UID]; exists {
		return fmt.Errorf("builtin plugin already registered: %s", def.Plugin.UID)
	}
	builtinCatalog.builtins[def.Plugin.UID] = builtin
	return nil
}

// MustRegisterBuiltin 注册内置插件，遇错 panic
func MustRegisterBuiltin(builtin Builtin) {
	if err := RegisterBuiltin(builtin); err != nil {
		panic(err)
	}
}

// Builtins 获取所有内置插件
func Builtins() []Builtin {
	builtinCatalog.RLock()
	defer builtinCatalog.RUnlock()

	items := make([]Builtin, 0, len(builtinCatalog.builtins))
	for _, builtin := range builtinCatalog.builtins {
		items = append(items, builtin)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Definition().Plugin.UID < items[j].Definition().Plugin.UID
	})
	return items
}

// FindBuiltin 查询内置插件
func FindBuiltin(pluginID string) (Builtin, bool) {
	builtinCatalog.RLock()
	defer builtinCatalog.RUnlock()

	builtin, ok := builtinCatalog.builtins[pluginID]
	return builtin, ok
}
