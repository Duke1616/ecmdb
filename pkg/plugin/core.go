package plugin

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
)

func NormalizeModelUID(modelUID string) (string, error) {
	modelUID = strings.TrimSpace(modelUID)
	if modelUID == "" {
		return "", fmt.Errorf("model_uid 不能为空")
	}
	return modelUID, nil
}

func ValidateResolveRequest(req ResolveRequest) error {
	if req.PluginID == "" {
		return fmt.Errorf("plugin_id 不能为空")
	}
	if req.Action == "" {
		return fmt.Errorf("action 不能为空")
	}
	return ValidateResourceID(req.ResourceID)
}

func ValidateResourceID(resourceID int64) error {
	if resourceID <= 0 {
		return fmt.Errorf("resource_id 参数错误")
	}
	return nil
}

func PrepareBindings(pluginID string, bindings []Binding) ([]Binding, error) {
	if len(bindings) == 0 {
		return nil, fmt.Errorf("bindings 不能为空")
	}

	prepared := make([]Binding, 0, len(bindings))
	for _, binding := range bindings {
		binding.PluginID = pluginID
		binding, err := PrepareBinding(binding)
		if err != nil {
			return nil, err
		}
		prepared = append(prepared, binding)
	}
	return prepared, nil
}

func PrepareBinding(binding Binding) (Binding, error) {
	var err error
	binding, err = normalizeBindingGraph(binding)
	if err != nil {
		return Binding{}, err
	}
	if err = binding.Validate(); err != nil {
		return Binding{}, err
	}
	return binding, nil
}

func (p Plugin) Validate() error {
	if p.UID == "" {
		return fmt.Errorf("插件 UID 不能为空")
	}
	if p.Name == "" {
		return fmt.Errorf("插件名称不能为空")
	}
	return nil
}

func (p Plugin) FindAction(name string) (ActionSpec, bool) {
	return lo.Find(p.Actions, func(action ActionSpec) bool {
		return action.Action == name
	})
}

func (p Plugin) ResourceActions() []ResourceAction {
	return lo.Map(p.Actions, func(action ActionSpec, _ int) ResourceAction {
		return ResourceAction{
			PluginID:   p.UID,
			Action:     action.Action,
			Name:       action.Name,
			Icon:       action.Icon,
			Placement:  action.Placement,
			Permission: action.Permission,
			BindingUID: action.BindingUID,
			Runtime:    action.Runtime,
			Meta:       action.Meta,
		}
	})
}

func (b Binding) Validate() error {
	if b.UID == "" {
		return fmt.Errorf("插件绑定 UID 不能为空")
	}
	if b.PluginID == "" {
		return fmt.Errorf("plugin_id 不能为空")
	}
	if b.ModelUID == "" {
		return fmt.Errorf("model_uid 不能为空")
	}
	if b.Graph == nil || len(b.Graph.Nodes) == 0 {
		return fmt.Errorf("graph 不能为空")
	}
	return nil
}

func normalizeBindingGraph(binding Binding) (Binding, error) {
	if binding.Graph == nil || len(binding.Graph.Nodes) == 0 {
		return binding, nil
	}
	if _, err := CompileBindingGraph(binding.Graph); err != nil {
		return Binding{}, err
	}
	if entry, ok := GraphEntryNode(binding.Graph); ok {
		binding.ModelUID = strings.TrimSpace(entry.ModelUID)
	}
	return binding, nil
}
