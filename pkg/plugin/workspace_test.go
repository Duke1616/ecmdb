package plugin

import (
	"testing"
)

func TestWorkspacePresetBuilder(t *testing.T) {
	reg := NewRegistry("builtin.ssh", "SSH").
		Action("terminal", "SSH 终端",
			Icon("terminal"),
			Workspace("Web Shell", "host",
				CardFields("name", "ip"),
				Prop("connectionType", "Web Shell"),
				Prop("autoConnect", true),
			),
		)

	def := reg.MustDefinition()
	if len(def.Plugin.Actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(def.Plugin.Actions))
	}

	action := def.Plugin.Actions[0]
	if action.Runtime == nil {
		t.Fatal("expected action runtime to be set")
	}
	if action.Runtime.Layout != "workspace" {
		t.Errorf("expected layout 'workspace', got '%s'", action.Runtime.Layout)
	}
	if action.Runtime.Title != "Web Shell" {
		t.Errorf("expected title 'Web Shell', got '%s'", action.Runtime.Title)
	}
	if action.Runtime.Props["connectionType"] != "Web Shell" || action.Runtime.Props["autoConnect"] != true {
		t.Errorf("unexpected props: %#v", action.Runtime.Props)
	}

	sidebar := action.Runtime.Sidebar
	if sidebar == nil || sidebar.Enabled == nil || !*sidebar.Enabled {
		t.Fatal("expected sidebar to be enabled")
	}
	if sidebar.Title != "Web Shell" {
		t.Errorf("expected sidebar title 'Web Shell', got '%s'", sidebar.Title)
	}
	if sidebar.Resource == nil || sidebar.Resource.ModelUID != "host" {
		t.Fatalf("unexpected sidebar resource: %#v", sidebar.Resource)
	}
	if sidebar.Resource.TitleField != "name" || sidebar.Resource.SubtitleField != "ip" {
		t.Errorf("unexpected card fields: title=%s, sub=%s", sidebar.Resource.TitleField, sidebar.Resource.SubtitleField)
	}
	if sidebar.Resource.Limit != 20 {
		t.Errorf("expected default limit 20, got %d", sidebar.Resource.Limit)
	}
}
