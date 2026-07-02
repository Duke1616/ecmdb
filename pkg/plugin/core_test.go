package plugin

import "testing"

func TestPrepareBindingDerivesModelUIDFromEntryNode(t *testing.T) {
	graph, err := GraphFromBindingSpecs("host", []ResourceSpec{
		{
			Name:        "target",
			ModelUID:    "host",
			Cardinality: CardinalityOne,
			Required:    true,
		},
	})
	if err != nil {
		t.Fatalf("GraphFromBindingSpecs() error = %v", err)
	}

	binding, err := PrepareBinding(Binding{
		UID:      "builtin.ssh.host",
		PluginID: "builtin.ssh",
		Graph:    graph,
	})
	if err != nil {
		t.Fatalf("PrepareBinding() error = %v", err)
	}

	if binding.ModelUID != "host" {
		t.Fatalf("expected model_uid host, got %s", binding.ModelUID)
	}
}

func TestPluginFindAction(t *testing.T) {
	plugin := Plugin{
		UID: "builtin.ssh",
		Actions: []ActionSpec{
			{Action: "terminal", Name: "SSH"},
		},
	}

	action, ok := plugin.FindAction("terminal")
	if !ok {
		t.Fatal("expected action to be found")
	}
	if action.Name != "SSH" {
		t.Fatalf("expected action name SSH, got %s", action.Name)
	}
}

func TestPluginResourceActions(t *testing.T) {
	plugin := Plugin{
		UID: "builtin.ssh",
		Actions: []ActionSpec{
			{Action: "terminal", Name: "SSH", UI: UIBuiltinTerminal},
		},
	}

	actions := plugin.ResourceActions()
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if actions[0].PluginID != "builtin.ssh" {
		t.Fatalf("expected plugin id builtin.ssh, got %s", actions[0].PluginID)
	}
}
