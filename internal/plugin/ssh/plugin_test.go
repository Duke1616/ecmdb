package ssh

import (
	"testing"

	"github.com/Duke1616/ecmdb/pkg/plugin"
)

func TestDefinitionProvidesHostBinding(t *testing.T) {
	def := Definition()
	if len(def.Bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(def.Bindings))
	}

	binding := def.Bindings[0]
	if binding.ModelUID != "host" {
		t.Fatalf("expected host binding, got %s", binding.ModelUID)
	}
	if binding.Graph == nil {
		t.Fatalf("binding %s graph is nil", binding.UID)
	}
	if binding.Graph.EntryNodeID != inputEndpoint {
		t.Fatalf("binding %s entry node = %s", binding.UID, binding.Graph.EntryNodeID)
	}
	specs, err := plugin.CompileBindingGraph(binding.Graph)
	if err != nil {
		t.Fatalf("binding %s compile error = %v", binding.UID, err)
	}
	if len(specs) != 1 || specs[0].Name != inputEndpoint {
		t.Fatalf("binding %s specs = %#v", binding.UID, specs)
	}
}

func TestDecodeTargetUsesRootInput(t *testing.T) {
	actionCtx := plugin.ActionContext{
		Binding: plugin.Binding{
			Graph: &plugin.BindingGraph{
				EntryNodeID: inputEndpoint,
				Nodes: []plugin.BindingGraphNode{
					{
						ID:          inputEndpoint,
						Name:        inputEndpoint,
						ModelUID:    "host",
						Cardinality: plugin.CardinalityOne,
						Required:    true,
					},
				},
			},
		},
		Inputs: map[string]plugin.ResolvedInput{
			inputEndpoint: {
				Name:        inputEndpoint,
				Cardinality: plugin.CardinalityOne,
				Resources: []plugin.ResolvedResource{
					{
						Fields: map[string]any{
							"host":     "10.0.0.8",
							"port":     22,
							"username": "root",
						},
					},
				},
			},
		},
	}

	target, err := DecodeTarget(actionCtx)
	if err != nil {
		t.Fatalf("DecodeTarget() error = %v", err)
	}
	if target.Host != "10.0.0.8" || target.Port != 22 || target.Username != "root" {
		t.Fatalf("unexpected target: %#v", target)
	}
}
