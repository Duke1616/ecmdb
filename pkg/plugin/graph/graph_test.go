package graph

import (
	"testing"

	"github.com/Duke1616/ecmdb/pkg/plugin/types"
)

func TestCompileBindingGraphRoundTrip(t *testing.T) {
	origSpecs := []types.ResourceSpec{
		{
			Name:        "target",
			ModelUID:    "host",
			Cardinality: types.CardinalityOne,
			Required:    true,
			Fields:      map[string]string{"ip": "ip"},
			Children: []types.ResourceSpec{
				{
					Name:         "gateways",
					ModelUID:     "gateway",
					RelationType: types.RelationTypeDefault,
					Direction:    types.DirectionToTarget,
					Cardinality:  types.CardinalityMany,
					Fields:       map[string]string{"host": "host"},
				},
			},
		},
	}

	graph, err := GraphFromBindingSpecs("host", origSpecs)
	if err != nil {
		t.Fatalf("GraphFromBindingSpecs() error = %v", err)
	}

	compiled, err := CompileBindingGraph(graph)
	if err != nil {
		t.Fatalf("CompileBindingGraph() error = %v", err)
	}

	if len(compiled) != 1 {
		t.Fatalf("expected 1 root spec, got %d", len(compiled))
	}
	root := compiled[0]
	if root.ModelUID != "host" || root.Name != "target" {
		t.Errorf("unexpected root: %#v", root)
	}
	if len(root.Children) != 1 || root.Children[0].ModelUID != "gateway" {
		t.Errorf("unexpected children: %#v", root.Children)
	}
}
