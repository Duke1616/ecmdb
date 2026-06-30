package plugin

type ModelBuilder struct {
	model ModelSpec
}

type Attr interface {
	Build() Attribute
}

type AttributeBuilder struct {
	attr Attribute
}

func Model(uid string, name string, opts ...ModelOption) *ModelBuilder {
	model := ModelSpec{
		UID:     uid,
		Name:    name,
		Builtin: true,
	}
	for _, opt := range opts {
		opt(&model)
	}
	return &ModelBuilder{model: model}
}

func (b *ModelBuilder) group(name string, index int64, fields ...Attribute) *ModelBuilder {
	b.model.AttributeGroups = append(b.model.AttributeGroups, AttributeGroup{
		Name:   name,
		Index:  index,
		Fields: fields,
	})
	return b
}

func (b *ModelBuilder) AttrGroup(name string, index int64, attrs ...Attr) *ModelBuilder {
	fields := make([]Attribute, 0, len(attrs))
	for _, attr := range attrs {
		if attr != nil {
			fields = append(fields, attr.Build())
		}
	}
	return b.group(name, index, fields...)
}

func (b *ModelBuilder) Build() ModelSpec {
	return b.model
}

func (b *ModelBuilder) applyToSchema(schema *Schema) {
	schema.Models = append(schema.Models, b.Build())
}

func (m ModelSpec) applyToSchema(schema *Schema) {
	schema.Models = append(schema.Models, m)
}

type ModelOption func(*ModelSpec)

func ModelIcon(icon string) ModelOption {
	return func(model *ModelSpec) {
		model.Icon = icon
	}
}

func ModelGroupName(groupName string) ModelOption {
	return func(model *ModelSpec) {
		model.GroupName = groupName
	}
}

func ModelBuiltin(builtin bool) ModelOption {
	return func(model *ModelSpec) {
		model.Builtin = builtin
	}
}

type AttributeOption func(*Attribute)

func Field(uid string, name string, fieldType string, opts ...AttributeOption) Attribute {
	attr := Attribute{
		UID:     uid,
		Name:    name,
		Type:    fieldType,
		Builtin: true,
	}
	for _, opt := range opts {
		opt(&attr)
	}
	return attr
}

func String(uid string, name string, opts ...AttributeOption) *AttributeBuilder {
	return newAttribute(uid, name, "string", opts...)
}

func List(uid string, name string, option any, opts ...AttributeOption) *AttributeBuilder {
	return newAttribute(uid, name, "list", opts...).Options(option)
}

func Multiline(uid string, name string, opts ...AttributeOption) *AttributeBuilder {
	return newAttribute(uid, name, "multiline", opts...)
}

func newAttribute(uid string, name string, fieldType string, opts ...AttributeOption) *AttributeBuilder {
	return &AttributeBuilder{attr: Field(uid, name, fieldType, opts...)}
}

func (b *AttributeBuilder) Build() Attribute {
	return b.attr
}

func (b *AttributeBuilder) Required() *AttributeBuilder {
	b.attr.Required = true
	return b
}

func (b *AttributeBuilder) Display() *AttributeBuilder {
	b.attr.Display = true
	return b
}

func (b *AttributeBuilder) Secure() *AttributeBuilder {
	b.attr.Secure = true
	return b
}

func (b *AttributeBuilder) Builtin(builtin bool) *AttributeBuilder {
	b.attr.Builtin = builtin
	return b
}

func (b *AttributeBuilder) Index(index int64) *AttributeBuilder {
	b.attr.Index = index
	return b
}

func (b *AttributeBuilder) Options(option any) *AttributeBuilder {
	b.attr.Option = option
	return b
}

func AttributeBuiltin(builtin bool) AttributeOption {
	return func(attr *Attribute) {
		attr.Builtin = builtin
	}
}
