package plugin

import (
	"testing"

	pluginx "github.com/Duke1616/ecmdb/pkg/plugin"
)

func TestBuildImportSchema(t *testing.T) {
	builtin := pluginx.StaticBuiltin(pluginx.Definition{
		Schema: pluginx.Schema{
			ModelGroups: []pluginx.ModelGroupSpec{
				{Name: "主机模型"},
				{Name: "网关模型"},
			},
			Models: []pluginx.ModelSpec{
				{UID: "host", GroupName: "主机模型"},
				{UID: "AuthGateway", GroupName: "网关模型"},
			},
			RelationTypes: []pluginx.RelationType{
				{UID: pluginx.RelationTypeDefault},
				{UID: pluginx.RelationTypeRun},
			},
			ModelRelations: []pluginx.ModelRelation{
				{
					SourceModelUID:  "AuthGateway",
					TargetModelUID:  "host",
					RelationTypeUID: pluginx.RelationTypeDefault,
					Mapping:         pluginx.MappingOneToMany,
				},
			},
		},
	})

	t.Run("only import referenced gateway schema", func(t *testing.T) {
		got, err := builtin.SchemaForBindings([]pluginx.Binding{
			{
				ModelUID: "server",
				Graph: &pluginx.BindingGraph{
					EntryNodeID: "entry",
					Edges: []pluginx.BindingGraphEdge{
						{
							From:         "entry",
							To:           "gateway",
							RelationType: pluginx.RelationTypeDefault,
							Direction:    pluginx.DirectionToSource,
						},
					},
					Nodes: []pluginx.BindingGraphNode{
						{ID: "entry", ModelUID: "server", Cardinality: pluginx.CardinalityOne},
						{ID: "gateway", ModelUID: "AuthGateway", Cardinality: pluginx.CardinalityMany},
					},
				},
			},
		})
		if err != nil {
			t.Fatalf("buildImportSchema() error = %v", err)
		}

		if len(got.Models) != 1 || got.Models[0].UID != "AuthGateway" {
			t.Fatalf("expected only AuthGateway model, got %+v", got.Models)
		}
		if len(got.ModelGroups) != 1 || got.ModelGroups[0].Name != "网关模型" {
			t.Fatalf("expected only 网关模型 group, got %+v", got.ModelGroups)
		}
		if len(got.ModelRelations) != 1 {
			t.Fatalf("expected 1 relation, got %+v", got.ModelRelations)
		}
		if relation := got.ModelRelations[0]; relation.SourceModelUID != "AuthGateway" || relation.TargetModelUID != "server" {
			t.Fatalf("expected relation AuthGateway -> server, got %+v", relation)
		}
		if got.ModelRelations[0].Mapping != pluginx.MappingManyToMany {
			t.Fatalf("expected many_to_many mapping, got %+v", got.ModelRelations[0])
		}
		if len(got.RelationTypes) != 1 || got.RelationTypes[0].UID != pluginx.RelationTypeDefault {
			t.Fatalf("expected default relation type only, got %+v", got.RelationTypes)
		}
	})

	t.Run("import full linked schema when both models referenced", func(t *testing.T) {
		got, err := builtin.SchemaForBindings([]pluginx.Binding{
			{
				ModelUID: "host",
				Graph: &pluginx.BindingGraph{
					EntryNodeID: "entry",
					Edges: []pluginx.BindingGraphEdge{
						{
							From:         "entry",
							To:           "gateway",
							RelationType: pluginx.RelationTypeDefault,
							Direction:    pluginx.DirectionToSource,
						},
					},
					Nodes: []pluginx.BindingGraphNode{
						{ID: "entry", ModelUID: "host", Cardinality: pluginx.CardinalityOne},
						{ID: "gateway", ModelUID: "AuthGateway", Cardinality: pluginx.CardinalityMany},
					},
				},
			},
		})
		if err != nil {
			t.Fatalf("buildImportSchema() error = %v", err)
		}

		if len(got.Models) != 2 {
			t.Fatalf("expected 2 models, got %+v", got.Models)
		}
		if len(got.ModelRelations) != 1 {
			t.Fatalf("expected 1 relation, got %+v", got.ModelRelations)
		}
		if len(got.RelationTypes) != 1 || got.RelationTypes[0].UID != pluginx.RelationTypeDefault {
			t.Fatalf("expected default relation type only, got %+v", got.RelationTypes)
		}
		if got.ModelRelations[0].Mapping != pluginx.MappingManyToMany {
			t.Fatalf("expected many_to_many mapping, got %+v", got.ModelRelations[0])
		}
	})
}
