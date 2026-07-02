package plugin

import "testing"

func TestRegistryDefinition(t *testing.T) {
	def := NewRegistry(
		"builtin.test",
		"Test",
		Type("builtin"),
	).
		Action("run", "运行", Icon("play"), UI("builtin:test")).
		Bind(Center[testHost]("host")).
		MustDefinition()

	if def.Plugin.UID != "builtin.test" || def.Plugin.Type != "builtin" {
		t.Fatalf("plugin = %#v", def.Plugin)
	}
	if len(def.Plugin.Actions) != 1 || def.Plugin.Actions[0].Action != "run" {
		t.Fatalf("actions = %#v", def.Plugin.Actions)
	}
	if len(def.Bindings) != 1 {
		t.Fatalf("bindings = %#v", def.Bindings)
	}
	binding := def.Bindings[0]
	if binding.PluginID != "builtin.test" || binding.ModelUID != "host" {
		t.Fatalf("binding = %#v", binding)
	}
	if binding.Graph == nil {
		t.Fatalf("binding graph is nil")
	}
	specs, err := CompileBindingGraph(binding.Graph)
	if err != nil {
		t.Fatalf("CompileBindingGraph() error = %v", err)
	}
	if specs[0].ModelUID != "host" {
		t.Fatalf("top spec model = %s", specs[0].ModelUID)
	}
	if specs[0].Children[0].RelationType != RelationTypeDefault {
		t.Fatalf("child relation type = %s", specs[0].Children[0].RelationType)
	}
}

func TestRegistryDefinitionWithCenter(t *testing.T) {
	def := NewRegistry("builtin.center", "Center").
		Bind(Center[testHost]("host")).
		MustDefinition()

	if len(def.Bindings) != 1 {
		t.Fatalf("bindings = %#v", def.Bindings)
	}
	if def.Bindings[0].UID != "builtin.center.host" {
		t.Fatalf("binding uid = %s", def.Bindings[0].UID)
	}
	specs, err := CompileBindingGraph(def.Bindings[0].Graph)
	if err != nil {
		t.Fatalf("CompileBindingGraph() error = %v", err)
	}
	spec := specs[0]
	if spec.Name != "target" || spec.ModelUID != "host" {
		t.Fatalf("spec = %#v", spec)
	}
	if spec.Children[0].RelationType != RelationTypeDefault {
		t.Fatalf("relation type = %s", spec.Children[0].RelationType)
	}
}

func TestRegistryDefinitionWithSchema(t *testing.T) {
	def := NewRegistry("builtin.schema", "Schema").
		Setup(
			ModelGroup("主机模型"),
			RelationTypes(BasicRelationTypes()...),
			Model("host", "主机", ModelGroupName("主机模型")).
				AttrGroup("基础属性", 0,
					String("ip", "IP地址").Required().Display().Index(1),
				),
			Relation("gateway", RelationTypeDefault, "host").OneToMany(),
		).
		MustDefinition()

	if len(def.Schema.ModelGroups) != 1 || def.Schema.ModelGroups[0].Name != "主机模型" {
		t.Fatalf("model groups = %#v", def.Schema.ModelGroups)
	}
	if len(def.Schema.RelationTypes) != 4 {
		t.Fatalf("relation types = %#v", def.Schema.RelationTypes)
	}
	if len(def.Schema.Models) != 1 || def.Schema.Models[0].UID != "host" {
		t.Fatalf("models = %#v", def.Schema.Models)
	}
	if len(def.Schema.Models[0].AttributeGroups) != 1 {
		t.Fatalf("attribute groups = %#v", def.Schema.Models[0].AttributeGroups)
	}
	field := def.Schema.Models[0].AttributeGroups[0].Fields[0]
	if field.UID != "ip" || !field.Required || !field.Display || field.Index != 1 {
		t.Fatalf("field = %#v", field)
	}
	if len(def.Schema.ModelRelations) != 1 || def.Schema.ModelRelations[0].Mapping != MappingOneToMany {
		t.Fatalf("model relations = %#v", def.Schema.ModelRelations)
	}
}
