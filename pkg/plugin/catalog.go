package plugin

import (
	"fmt"
	"sort"
	"sync"
)

var builtinCatalog = struct {
	sync.RWMutex
	builtins map[string]Builtin
}{
	builtins: make(map[string]Builtin),
}

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

func MustRegisterBuiltin(builtin Builtin) {
	if err := RegisterBuiltin(builtin); err != nil {
		panic(err)
	}
}

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

func FindBuiltin(pluginID string) (Builtin, bool) {
	builtinCatalog.RLock()
	defer builtinCatalog.RUnlock()

	builtin, ok := builtinCatalog.builtins[pluginID]
	return builtin, ok
}
