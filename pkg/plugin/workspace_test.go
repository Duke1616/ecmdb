package plugin

import (
	"testing"

	"github.com/Duke1616/ecmdb/pkg/plugin/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspace(t *testing.T) {
	testCases := []struct {
		name       string
		title      string
		modelUID   string
		opts       []WorkspaceOption
		assertSpec func(t *testing.T, action types.ActionSpec)
	}{
		{
			name:     "默认工作台预设配置",
			title:    "Web Shell",
			modelUID: "host",
			opts: []WorkspaceOption{
				CardFields("name", "ip"),
				Prop("autoConnect", true),
			},
			assertSpec: func(t *testing.T, action types.ActionSpec) {
				require.NotNil(t, action.Runtime)
				assert.Equal(t, "workspace", action.Runtime.Layout)
				assert.Equal(t, "Web Shell", action.Runtime.Title)
				assert.Equal(t, true, action.Runtime.Props["autoConnect"])

				sidebar := action.Runtime.Sidebar
				require.NotNil(t, sidebar)
				assert.True(t, *sidebar.Enabled)
				assert.Equal(t, "Web Shell", sidebar.Title)
				require.NotNil(t, sidebar.Resource)
				assert.Equal(t, "host", sidebar.Resource.ModelUID)
				assert.Equal(t, "name", sidebar.Resource.TitleField)
				assert.Equal(t, "ip", sidebar.Resource.SubtitleField)
				assert.Equal(t, 20, sidebar.Resource.Limit)
			},
		},
		{
			name:     "自定义侧边栏标题与单页限制",
			title:    "文件管理器",
			modelUID: "host",
			opts: []WorkspaceOption{
				SidebarTitle("资产导航"),
				SidebarLimit(50),
				SearchFields("name", "ip", "hostname"),
			},
			assertSpec: func(t *testing.T, action types.ActionSpec) {
				sidebar := action.Runtime.Sidebar
				require.NotNil(t, sidebar)
				assert.Equal(t, "资产导航", sidebar.Title)
				assert.Equal(t, 50, sidebar.Resource.Limit)
				assert.Equal(t, []string{"name", "ip", "hostname"}, sidebar.Resource.SearchFields)
			},
		},
		{
			name:     "禁用侧边栏资产导航",
			title:    "独立看板",
			modelUID: "host",
			opts: []WorkspaceOption{
				SidebarDisabled(),
				Props(map[string]any{"theme": "dark", "zoom": 1.2}),
			},
			assertSpec: func(t *testing.T, action types.ActionSpec) {
				sidebar := action.Runtime.Sidebar
				require.NotNil(t, sidebar)
				assert.False(t, *sidebar.Enabled)
				assert.Equal(t, "dark", action.Runtime.Props["theme"])
				assert.Equal(t, 1.2, action.Runtime.Props["zoom"])
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			spec := types.ActionSpec{
				Action: "test_action",
				Name:   tc.title,
			}
			opt := Workspace(tc.title, tc.modelUID, tc.opts...)
			opt(&spec)
			tc.assertSpec(t, spec)
		})
	}
}
