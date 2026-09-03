package web

import (
	"testing"

	"github.com/Duke1616/ecmdb/internal/domain"
	pluginx "github.com/Duke1616/ecmdb/pkg/plugin/types"
)

func TestBuildRuntimeViewFromSpecUsesTypedActionRuntime(t *testing.T) {
	p := domain.Plugin{
		UID:     "builtin.ssh",
		Name:    "SSH",
		Version: "1.0.0",
	}
	spec := pluginx.ActionSpec{
		Action: "terminal",
		Name:   "SSH 终端",
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
	}
	view := buildRuntimeViewFromSpec(p, spec, 42)

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

func TestBuildRuntimeViewFromSpecDefaultsToActionName(t *testing.T) {
	p := domain.Plugin{
		UID:     "builtin.ssh",
		Name:    "SSH",
		Version: "1.0.0",
	}
	spec := pluginx.ActionSpec{
		Action: "terminal",
		Name:   "SSH 终端",
	}
	view := buildRuntimeViewFromSpec(p, spec, 42)

	if view.Presentation.Title != "SSH 终端" {
		t.Fatalf("expected title to default to action name, got %s", view.Presentation.Title)
	}
	if got := view.Runtime.Props["resourceId"]; got != "42" {
		t.Fatalf("expected resourceId prop to be 42, got %v", got)
	}
}

func boolPtr(v bool) *bool {
	return &v
}
