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
			Enabled: true,
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

func Enabled(value bool) Option {
	return func(p *Plugin) {
		p.Enabled = value
	}
}

type ActionOption func(*ActionSpec)

func Icon(value string) ActionOption {
	return func(a *ActionSpec) {
		a.Icon = value
	}
}

func UI(value string) ActionOption {
	return func(a *ActionSpec) {
		a.UI = value
	}
}

func Placement(value string) ActionOption {
	return func(a *ActionSpec) {
		a.Placement = value
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
		b.Specs = Configure(b.Specs).ForPath(path, fn).Build()
	}
}

func At(path string, opts ...SpecOption) BindingOption {
	return Spec(path, func(spec *ResourceSpec) {
		for _, opt := range opts {
			opt(spec)
		}
	})
}

type SpecOption func(*ResourceSpec)

func In(relationType string) SpecOption {
	return func(spec *ResourceSpec) {
		spec.Direction = DirectionToSource
		spec.RelationType = relationType
	}
}

func Out(relationType string) SpecOption {
	return func(spec *ResourceSpec) {
		spec.Direction = DirectionToTarget
		spec.RelationType = relationType
	}
}

func Where(field string, operator string, value any) SpecOption {
	return func(spec *ResourceSpec) {
		spec.Filters = append(spec.Filters, Filter{
			Field:    field,
			Operator: operator,
			Value:    value,
		})
	}
}

func Required(value bool) SpecOption {
	return func(spec *ResourceSpec) {
		spec.Required = value
	}
}

func Center[T any](modelUID string, opts ...BindingOption) BindingFactory {
	return func(pluginUID string) (Binding, error) {
		uid := fmt.Sprintf("%s.%s", pluginUID, modelUID)
		return bindingForCenter[T](uid, modelUID, opts...)
	}
}

func bindingForCenter[T any](uid string, modelUID string, opts ...BindingOption) (Binding, error) {
	spec, err := BuildCenterSpec[T]("target", modelUID)
	if err != nil {
		return Binding{}, err
	}

	binding := Binding{
		UID:      uid,
		ModelUID: modelUID,
		Enabled:  true,
		Specs:    []ResourceSpec{spec},
	}
	for _, opt := range opts {
		opt(&binding)
	}
	if err = validateResourceSpecs(binding.Specs); err != nil {
		return Binding{}, err
	}
	return binding, nil
}

func validateResourceSpecs(specs []ResourceSpec) error {
	for _, spec := range specs {
		if spec.RelationType != "" && !ValidRelationType(spec.RelationType) {
			return fmt.Errorf("unsupported relation type %s on %s", spec.RelationType, spec.Name)
		}
		if err := validateResourceSpecs(spec.Children); err != nil {
			return err
		}
	}
	return nil
}
