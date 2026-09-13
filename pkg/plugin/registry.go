package plugin

import (
	"fmt"

	"github.com/Duke1616/ecmdb/pkg/plugin/types"
)

// Definition 插件自描述元数据规范
type Definition struct {
	Plugin   types.Plugin    `json:"plugin"`
	Schema   types.Schema    `json:"schema"`
	Bindings []types.Binding `json:"bindings"`
}

// Registry 插件注册中心顶层建造者
type Registry struct {
	plugin   types.Plugin
	schema   types.Schema
	bindings []types.Binding
	err      error
}

// NewRegistry 创建插件注册表
func NewRegistry(uid string, name string, opts ...Option) *Registry {
	r := &Registry{
		plugin: types.Plugin{
			UID:     uid,
			Name:    name,
			Type:    "custom",
			Version: "1.0.0",
		},
	}
	for _, opt := range opts {
		opt(&r.plugin)
	}
	return r
}

// Action 注册无资产绑定的普通独立动作（如纯操作按钮、快捷动作等）
func (r *Registry) Action(action string, name string, opts ...ActionOption) *Registry {
	spec := types.ActionSpec{
		Action:    action,
		Name:      name,
		Placement: types.PlacementResourceDetailActions,
	}
	for _, opt := range opts {
		opt(&spec)
	}
	r.plugin.Actions = append(r.plugin.Actions, spec)
	return r
}

// Definition 编译并产出最终的插件自描述元数据
func (r *Registry) Definition() (Definition, error) {
	if r.err != nil {
		return Definition{}, r.err
	}
	if r.plugin.UID == "" {
		return Definition{}, fmt.Errorf("plugin uid is required")
	}
	if r.plugin.Name == "" {
		return Definition{}, fmt.Errorf("plugin name is required")
	}

	return Definition{
		Plugin:   r.plugin,
		Schema:   r.schema,
		Bindings: r.bindings,
	}, nil
}

// MustDefinition 编译并产出最终元数据，遇错 panic（常用于单元测试或常量定义）
func (r *Registry) MustDefinition() Definition {
	def, err := r.Definition()
	if err != nil {
		panic(err)
	}
	return def
}

// Option 插件基础属性修饰选项
type Option func(*types.Plugin)

func Type(value string) Option {
	return func(p *types.Plugin) {
		p.Type = value
	}
}

func Version(value string) Option {
	return func(p *types.Plugin) {
		p.Version = value
	}
}

// ActionOption 动作修饰选项
type ActionOption func(*types.ActionSpec)

func Icon(value string) ActionOption {
	return func(a *types.ActionSpec) {
		a.Icon = value
	}
}

func Placement(value string) ActionOption {
	return func(a *types.ActionSpec) {
		a.Placement = value
	}
}

func Permission(value string) ActionOption {
	return func(a *types.ActionSpec) {
		a.Permission = value
	}
}

func ActionRuntime(value types.ActionRuntimeSpec) ActionOption {
	return func(a *types.ActionSpec) {
		runtime := value
		a.Runtime = &runtime
	}
}

func Meta(key string, value any) ActionOption {
	return func(a *types.ActionSpec) {
		if a.Meta == nil {
			a.Meta = make(map[string]any)
		}
		a.Meta[key] = value
	}
}
