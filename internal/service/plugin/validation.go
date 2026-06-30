package plugin

import (
	"fmt"

	pluginx "github.com/Duke1616/ecmdb/pkg/plugin"
)

func validatePlugin(p pluginx.Plugin) error {
	if p.UID == "" {
		return fmt.Errorf("插件 UID 不能为空")
	}
	if p.Name == "" {
		return fmt.Errorf("插件名称不能为空")
	}
	return nil
}

func validateBinding(b pluginx.Binding) error {
	if b.UID == "" {
		return fmt.Errorf("插件绑定 UID 不能为空")
	}
	if b.PluginID == "" {
		return fmt.Errorf("plugin_id 不能为空")
	}
	if b.ModelUID == "" {
		return fmt.Errorf("model_uid 不能为空")
	}
	if len(b.Specs) == 0 {
		return fmt.Errorf("specs 不能为空")
	}
	return nil
}

func validateResolveRequest(req pluginx.ResolveRequest) error {
	if req.PluginID == "" {
		return fmt.Errorf("plugin_id 不能为空")
	}
	if req.Action == "" {
		return fmt.Errorf("action 不能为空")
	}
	if req.ResourceID <= 0 {
		return fmt.Errorf("resource_id 参数错误")
	}
	return nil
}
