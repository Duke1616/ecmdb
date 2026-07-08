package web

import (
	"testing"

	pluginx "github.com/Duke1616/ecmdb/pkg/plugin"
)

func TestBuildRuntimeViewUsesTypedActionRuntime(t *testing.T) {
	view := buildRuntimeView(pluginx.ResolveResult{
		PluginID:      "builtin.ssh",
		PluginName:    "SSH",
		PluginVersion: "1.0.0",
		Action:        "terminal",
		ResourceID:    42,
		Runtime: &pluginx.ActionRuntimeSpec{
			Layout: "workspace",
			Title:  "SSH 终端",
			Props: map[string]any{
				"title":          "SSH 终端",
				"connectionType": "Web Shell",
				"autoConnect":    true,
			},
			Sidebar: &pluginx.RuntimeSidebarSpec{
				Enabled:           boolPtr(true),
				Mode:              "resource-list",
				Title:             "资源列表",
				SearchPlaceholder: "搜索资源名称或 ID",
				EmptyText:         "暂无资源数据",
				Collapsible:       boolPtr(true),
				Resource: &pluginx.RuntimeSidebarResourceSpec{
					ModelUID: "host",
					Limit:    20,
				},
			},
		},
	})

	if view.Presentation.Layout != "workspace" {
		t.Fatalf("unexpected layout: %s", view.Presentation.Layout)
	}
	if view.Presentation.Title != "SSH 终端" {
		t.Fatalf("unexpected title: %s", view.Presentation.Title)
	}
	if view.Presentation.Sidebar == nil {
		t.Fatal("expected typed runtime sidebar")
	}
	if view.Presentation.Sidebar.Resource == nil {
		t.Fatal("expected sidebar resource config")
	}
	if view.Presentation.Sidebar.Resource.ModelUID != "host" {
		t.Fatalf("unexpected model uid: %s", view.Presentation.Sidebar.Resource.ModelUID)
	}
	if got := view.Runtime.Props["connectionType"]; got != "Web Shell" {
		t.Fatalf("unexpected connectionType: %v", got)
	}
	if got := view.Runtime.Props["title"]; got != "SSH 终端" {
		t.Fatalf("unexpected title prop: %v", got)
	}
	if view.Entry.JSURL != "/api/cmdb/plugin-runtime/builtin.ssh/static/index.umd.js?v=1.0.0" {
		t.Fatalf("unexpected js url: %s", view.Entry.JSURL)
	}
	if view.Entry.CSSURL != "/api/cmdb/plugin-runtime/builtin.ssh/static/index.css?v=1.0.0" {
		t.Fatalf("unexpected css url: %s", view.Entry.CSSURL)
	}
}

func TestBuildRuntimeViewDefaultsToActionName(t *testing.T) {
	view := buildRuntimeView(pluginx.ResolveResult{
		PluginID:   "builtin.ssh",
		PluginName: "SSH",
		ActionName: "SSH 终端",
		Action:     "terminal",
		ResourceID: 42,
	})

	if view.Presentation.Title != "SSH 终端" {
		t.Fatalf("unexpected presentation title: %s", view.Presentation.Title)
	}
	if _, ok := view.Runtime.Props["title"]; ok {
		t.Fatal("did not expect implicit title prop")
	}
}

func boolPtr(value bool) *bool {
	return &value
}
