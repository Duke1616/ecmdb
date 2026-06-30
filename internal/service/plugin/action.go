package plugin

import (
	"github.com/Duke1616/ecmdb/internal/domain"
	pluginx "github.com/Duke1616/ecmdb/pkg/plugin"
	"github.com/samber/lo"
)

func toResourceAction(p domain.Plugin, action domain.PluginActionSpec) pluginx.ResourceAction {
	return pluginx.ResourceAction{
		PluginID:  p.UID,
		Action:    action.Action,
		Name:      action.Name,
		Icon:      action.Icon,
		Placement: action.Placement,
		UI:        action.UI,
		Meta:      action.Meta,
	}
}

func findAction(actions []domain.PluginActionSpec, name string) (domain.PluginActionSpec, bool) {
	return lo.Find(actions, func(action domain.PluginActionSpec) bool {
		return action.Action == name
	})
}
