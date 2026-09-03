package plugin

import (
	"context"
	"fmt"

	"github.com/Duke1616/ecmdb/pkg/plugin/codec"
	"github.com/Duke1616/ecmdb/pkg/plugin/dsl"
	"github.com/Duke1616/ecmdb/pkg/plugin/graph"
	"github.com/Duke1616/ecmdb/pkg/plugin/types"
)

type Definition struct {
	Plugin   types.Plugin    `json:"plugin"`
	Schema   types.Schema    `json:"schema"`
	Bindings []types.Binding `json:"bindings"`
}

type Store interface {
	UpsertPlugin(ctx context.Context, p types.Plugin) error
	UpsertBinding(ctx context.Context, b types.Binding) error
}

func (d Definition) Save(ctx context.Context, store Store) error {
	if err := store.UpsertPlugin(ctx, d.Plugin); err != nil {
		return err
	}
	for _, binding := range d.Bindings {
		if err := store.UpsertBinding(ctx, binding); err != nil {
			return err
		}
	}
	return nil
}

type Registry struct {
	plugin   types.Plugin
	schema   types.Schema
	bindings []types.Binding
	err      error
}

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

func (r *Registry) Setup(items ...dsl.SetupItem) *Registry {
	for _, item := range items {
		if item != nil {
			item.ApplyToSchema(&r.schema)
		}
	}
	return r
}

func (r *Registry) Bind(factory BindingFactory) *Registry {
	if r.err != nil {
		return r
	}
	binding, err := factory(r.plugin.UID)
	if err != nil {
		r.err = err
		return r
	}
	if binding.PluginID == "" {
		binding.PluginID = r.plugin.UID
	}
	r.bindings = append(r.bindings, binding)
	return r
}

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

func (r *Registry) MustDefinition() Definition {
	def, err := r.Definition()
	if err != nil {
		panic(err)
	}
	return def
}

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

func UseBinding(uid string) ActionOption {
	return func(a *types.ActionSpec) {
		a.BindingUID = uid
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

type BindingOption func(*types.Binding)

type BindingFactory func(pluginUID string) (types.Binding, error)

func BindingEnabled(value bool) BindingOption {
	return func(b *types.Binding) {
		b.Enabled = value
	}
}

func Center[T any](modelUID string, opts ...BindingOption) BindingFactory {
	return CenterNamed[T]("target", modelUID, opts...)
}

func CenterNamed[T any](name string, modelUID string, opts ...BindingOption) BindingFactory {
	return func(pluginUID string) (types.Binding, error) {
		uid := CenterBindingUID(pluginUID, modelUID)
		return bindingForCenter[T](uid, name, modelUID, opts...)
	}
}

func CenterBindingUID(pluginUID string, modelUID string) string {
	return fmt.Sprintf("%s.%s", pluginUID, modelUID)
}

func bindingForCenter[T any](uid string, name string, modelUID string, opts ...BindingOption) (types.Binding, error) {
	spec, err := codec.BuildCenterSpec[T](name, modelUID)
	if err != nil {
		return types.Binding{}, err
	}
	g, err := graph.GraphFromBindingSpecs(modelUID, []types.ResourceSpec{spec})
	if err != nil {
		return types.Binding{}, err
	}

	binding := types.Binding{
		UID:      uid,
		ModelUID: modelUID,
		Enabled:  true,
		Graph:    g,
	}
	for _, opt := range opts {
		opt(&binding)
	}
	if _, err = graph.CompileBindingGraph(binding.Graph); err != nil {
		return types.Binding{}, err
	}
	return binding, nil
}
