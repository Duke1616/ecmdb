package plugin

type Builtin interface {
	Definition() Definition
	SchemaForBindings(bindings []Binding) (Schema, error)
}

type staticBuiltin struct {
	def Definition
}

func StaticBuiltin(def Definition) Builtin {
	return staticBuiltin{def: def}
}

func (b staticBuiltin) Definition() Definition {
	return b.def
}

func (b staticBuiltin) SchemaForBindings(bindings []Binding) (Schema, error) {
	return buildImportSchema(b.def.Schema, bindings)
}
