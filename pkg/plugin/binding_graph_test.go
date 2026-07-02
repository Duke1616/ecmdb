package plugin

import "testing"

func TestMutateBindingGraphPath(t *testing.T) {
	graph, err := GraphFromBindingSpecs("host", []ResourceSpec{
		{
			Name:        "target",
			ModelUID:    "host",
			Cardinality: CardinalityOne,
			Required:    true,
			Children: []ResourceSpec{
				{
					Name:         "gateways",
					ModelUID:     "gateway",
					RelationType: RelationTypeDefault,
					Direction:    DirectionToTarget,
					Cardinality:  CardinalityMany,
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("GraphFromBindingSpecs() error = %v", err)
	}

	ok := MutateBindingGraphPath(graph, "target.gateways", func(node *BindingGraphNode, edge *BindingGraphEdge) {
		node.Required = true
		edge.Direction = DirectionToSource
	})
	if !ok {
		t.Fatal("expected path to be resolved")
	}

	specs, err := CompileBindingGraph(graph)
	if err != nil {
		t.Fatalf("CompileBindingGraph() error = %v", err)
	}

	child := specs[0].Children[0]
	if !child.Required {
		t.Fatalf("expected child required=true, got %#v", child)
	}
	if child.Direction != DirectionToSource {
		t.Fatalf("expected child direction=%s, got %s", DirectionToSource, child.Direction)
	}
}

func TestMutateBindingGraphPathRejectsInvalidPath(t *testing.T) {
	graph := &BindingGraph{
		EntryNodeID: "target",
		Nodes: []BindingGraphNode{
			{ID: "target", Name: "target", ModelUID: "host", Cardinality: CardinalityOne},
		},
	}

	if ok := MutateBindingGraphPath(graph, "target..gateways", func(node *BindingGraphNode, edge *BindingGraphEdge) {}); ok {
		t.Fatal("expected invalid path to be rejected")
	}
}
