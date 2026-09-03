package plugin

import (
	"testing"

	"github.com/Duke1616/ecmdb/pkg/plugin/dsl"
	"github.com/Duke1616/ecmdb/pkg/plugin/graph"
	"github.com/Duke1616/ecmdb/pkg/plugin/types"
)

type testGateway struct {
	IP   string `plugin:"ip,required"`
	Port int    `plugin:"port"`
}

type testHost struct {
	IP       string        `plugin:"ip,required"`
	Username string        `plugin:"username"`
	Gateways []testGateway `plugin:"gateways,model=gateway,out=default"`
}

func TestRegistryDefinition(t *testing.T) {
	def := NewRegistry(
		"builtin.test",
		"Test",
		Type("builtin"),
	).
		Action("run", "运行", Icon("play")).
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
	specs, err := graph.CompileBindingGraph(binding.Graph)
	if err != nil {
		t.Fatalf("CompileBindingGraph() error = %v", err)
	}
	if specs[0].ModelUID != "host" {
		t.Fatalf("top spec model = %s", specs[0].ModelUID)
	}
	if specs[0].Children[0].RelationType != types.RelationTypeDefault {
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
	specs, err := graph.CompileBindingGraph(def.Bindings[0].Graph)
	if err != nil {
		t.Fatalf("CompileBindingGraph() error = %v", err)
	}
	spec := specs[0]
	if spec.Name != "target" || spec.ModelUID != "host" {
		t.Fatalf("spec = %#v", spec)
	}
	if spec.Children[0].RelationType != types.RelationTypeDefault {
		t.Fatalf("relation type = %s", spec.Children[0].RelationType)
	}
}

func TestRegistryDefinitionWithSchema(t *testing.T) {
	def := NewRegistry("builtin.schema", "Schema").
		Setup(
			dsl.ModelGroup("主机模型"),
			dsl.RelationTypes(dsl.BasicRelationTypes()...),
			dsl.Relation("gateway", types.RelationTypeDefault, "host").OneToMany(),
		).
		MustDefinition()

	if len(def.Schema.ModelGroups) != 1 || def.Schema.ModelGroups[0].Name != "主机模型" {
		t.Fatalf("model groups = %#v", def.Schema.ModelGroups)
	}
	if len(def.Schema.RelationTypes) != 4 {
		t.Fatalf("relation types = %#v", def.Schema.RelationTypes)
	}
	if len(def.Schema.ModelRelations) != 1 || def.Schema.ModelRelations[0].Mapping != types.MappingOneToMany {
		t.Fatalf("model relations = %#v", def.Schema.ModelRelations)
	}
}
