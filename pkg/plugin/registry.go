package plugin

import (
	"context"
	"fmt"
)

type Definition struct {
	Plugin   Plugin    `json:"plugin"`
	Schema   Schema    `json:"schema"`
	Bindings []Binding `json:"bindings"`
}

type Store interface {
	UpsertPlugin(ctx context.Context, p Plugin) error
	UpsertBinding(ctx context.Context, b Binding) error
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
	plugin   Plugin
	schema   Schema
	bindings []Binding
	err      error
}

func NewRegistry(uid string, name string, opts ...Option) *Registry {
	r := &Registry{
		plugin: Plugin{
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
	spec := ActionSpec{
		Action:    action,
		Name:      name,
		Placement: PlacementResourceDetailActions,
	}
	for _, opt := range opts {
		opt(&spec)
	}
	r.plugin.Actions = append(r.plugin.Actions, spec)
	return r
}

func (r *Registry) Setup(items ...SetupItem) *Registry {
	for _, item := range items {
		if item != nil {
			item.applyToSchema(&r.schema)
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

type Option func(*Plugin)

func Type(value string) Option {
	return func(p *Plugin) {
		p.Type = value
	}
}

func Version(value string) Option {
	return func(p *Plugin) {
		p.Version = value
	}
}

type ActionOption func(*ActionSpec)

func Icon(value string) ActionOption {
	return func(a *ActionSpec) {
		a.Icon = value
	}
}

func Placement(value string) ActionOption {
	return func(a *ActionSpec) {
		a.Placement = value
	}
}

func Permission(value string) ActionOption {
	return func(a *ActionSpec) {
		a.Permission = value
	}
}

func UseBinding(uid string) ActionOption {
	return func(a *ActionSpec) {
		a.BindingUID = uid
	}
}

func ActionRuntime(value ActionRuntimeSpec) ActionOption {
	return func(a *ActionSpec) {
		runtime := value
		a.Runtime = &runtime
	}
}

func Meta(key string, value any) ActionOption {
	return func(a *ActionSpec) {
		if a.Meta == nil {
			a.Meta = make(map[string]any)
		}
		a.Meta[key] = value
	}
}

type BindingOption func(*Binding)

type BindingFactory func(pluginUID string) (Binding, error)

func BindingEnabled(value bool) BindingOption {
	return func(b *Binding) {
		b.Enabled = value
	}
}

func Spec(path string, fn func(*ResourceSpec)) BindingOption {
	return func(b *Binding) {
		if b.Graph == nil {
			return
		}
		specs, err := CompileBindingGraph(b.Graph)
		if err != nil {
			panic(err)
		}
		specs = Configure(specs).ForPath(path, fn).Build()
		graph, err := GraphFromBindingSpecs(b.ModelUID, specs)
		if err != nil {
			panic(err)
		}
		b.Graph = graph
	}
}

func At(path string, opts ...SpecOption) BindingOption {
	return Node(path, func(node *BindingGraphNode, edge *BindingGraphEdge) {
		for _, opt := range opts {
			opt(node, edge)
		}
	})
}

type SpecOption func(*BindingGraphNode, *BindingGraphEdge)

func In(relationType string) SpecOption {
	return func(node *BindingGraphNode, edge *BindingGraphEdge) {
		if edge == nil {
			return
		}
		edge.Direction = DirectionToSource
		edge.RelationType = relationType
	}
}

func Out(relationType string) SpecOption {
	return func(node *BindingGraphNode, edge *BindingGraphEdge) {
		if edge == nil {
			return
		}
		edge.Direction = DirectionToTarget
		edge.RelationType = relationType
	}
}

func Where(field string, operator string, value any) SpecOption {
	return func(node *BindingGraphNode, edge *BindingGraphEdge) {
		node.Filters = append(node.Filters, Filter{
			Field:    field,
			Operator: operator,
			Value:    value,
		})
	}
}

func Required(value bool) SpecOption {
	return func(node *BindingGraphNode, edge *BindingGraphEdge) {
		node.Required = value
	}
}

func Center[T any](modelUID string, opts ...BindingOption) BindingFactory {
	return CenterNamed[T]("target", modelUID, opts...)
}

func CenterNamed[T any](name string, modelUID string, opts ...BindingOption) BindingFactory {
	return func(pluginUID string) (Binding, error) {
		uid := CenterBindingUID(pluginUID, modelUID)
		return bindingForCenter[T](uid, name, modelUID, opts...)
	}
}

func CenterBindingUID(pluginUID string, modelUID string) string {
	return fmt.Sprintf("%s.%s", pluginUID, modelUID)
}

func bindingForCenter[T any](uid string, name string, modelUID string, opts ...BindingOption) (Binding, error) {
	graph, err := BuildCenterGraph[T](name, modelUID)
	if err != nil {
		return Binding{}, err
	}

	binding := Binding{
		UID:      uid,
		ModelUID: modelUID,
		Enabled:  true,
		Graph:    graph,
	}
	for _, opt := range opts {
		opt(&binding)
	}
	if _, err = CompileBindingGraph(binding.Graph); err != nil {
		return Binding{}, err
	}
	return binding, nil
}

func Node(path string, fn func(node *BindingGraphNode, edge *BindingGraphEdge)) BindingOption {
	return func(b *Binding) {
		if b.Graph == nil {
			return
		}
		MutateBindingGraphPath(b.Graph, path, fn)
	}
}
