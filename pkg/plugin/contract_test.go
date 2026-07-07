package plugin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExternalServiceRuntime(t *testing.T) {
	def := NewRegistry(
		"ssh",
		"SSH",
		ExternalServiceRuntime("http://ssh-plugin:8080/", RuntimeHealthPath("/healthz")),
	).
		Action("terminal", "SSH 终端", UI(UIBuiltinTerminal)).
		MustDefinition()

	runtime, ok := def.Plugin.Runtime()
	if !ok {
		t.Fatal("runtime not found")
	}
	if runtime.Mode != RuntimeModeExternalService {
		t.Fatalf("mode = %s", runtime.Mode)
	}
	if runtime.Upstream != "http://ssh-plugin:8080" {
		t.Fatalf("upstream = %s", runtime.Upstream)
	}
	if runtime.HealthPath != "/healthz" {
		t.Fatalf("health path = %s", runtime.HealthPath)
	}
}

func TestDefinitionHandler(t *testing.T) {
	provider := ProviderFunc(func() (Definition, error) {
		return NewRegistry("ssh", "SSH").
			Action("terminal", "SSH 终端", UI(UIBuiltinTerminal)).
			Definition()
	})

	req := httptest.NewRequest(http.MethodGet, WellKnownPath, nil)
	rec := httptest.NewRecorder()
	DefinitionHandler(provider).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var def Definition
	if err := json.Unmarshal(rec.Body.Bytes(), &def); err != nil {
		t.Fatalf("decode definition: %v", err)
	}
	if def.Plugin.UID != "ssh" || len(def.Plugin.Actions) != 1 {
		t.Fatalf("definition = %#v", def)
	}
}

func TestPluginSchemaAndBindingsMetaRoundTrip(t *testing.T) {
	plugin := Plugin{UID: "builtin.ssh", Name: "SSH"}
	schema := Schema{
		Models: []ModelSpec{
			{UID: "host", Name: "主机"},
		},
	}
	bindings := []Binding{
		{
			UID:      "builtin.ssh.host",
			PluginID: "builtin.ssh",
			ModelUID: "host",
			Enabled:  true,
			Graph: &BindingGraph{
				EntryNodeID: "target",
				Nodes: []BindingGraphNode{
					{ID: "target", Name: "target", ModelUID: "host", Cardinality: CardinalityOne, Required: true},
				},
			},
		},
	}

	plugin.SetSchema(schema)
	plugin.SetDefaultBindings(bindings)

	gotSchema, ok := plugin.Schema()
	if !ok {
		t.Fatal("expected schema in meta")
	}
	if len(gotSchema.Models) != 1 || gotSchema.Models[0].UID != "host" {
		t.Fatalf("unexpected schema: %#v", gotSchema)
	}

	gotBindings, ok := plugin.DefaultBindings()
	if !ok {
		t.Fatal("expected bindings in meta")
	}
	if len(gotBindings) != 1 || gotBindings[0].UID != "builtin.ssh.host" {
		t.Fatalf("unexpected bindings: %#v", gotBindings)
	}
}
