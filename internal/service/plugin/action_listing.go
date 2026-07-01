package plugin

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/domain"
	pluginx "github.com/Duke1616/ecmdb/pkg/plugin"
	"github.com/samber/lo"
)

type bindingActionMatcher func(binding domain.PluginBinding, plugin domain.Plugin) (bool, error)

func (s *service) ListResourceActions(ctx context.Context, resourceID int64) ([]pluginx.ResourceAction, error) {
	if err := validateResourceID(resourceID); err != nil {
		return nil, err
	}

	primary, err := s.resolver.loadResource(ctx, resourceID, nil)
	if err != nil {
		return nil, err
	}

	return s.listActionsForResource(ctx, primary)
}

func (s *service) ListModelActions(ctx context.Context, modelUID string) ([]pluginx.ResourceAction, error) {
	uid, err := normalizeModelUID(modelUID)
	if err != nil {
		return nil, err
	}

	bindings, err := s.repo.ListEnabledBindingsByModelUID(ctx, uid)
	if err != nil {
		return nil, err
	}

	return s.listActionsForModel(ctx, bindings, map[string]domain.Plugin{})
}

func (s *service) listActionsForResource(ctx context.Context, primary domain.Resource) ([]pluginx.ResourceAction, error) {
	bindings, err := s.repo.ListEnabledBindingsByModelUID(ctx, primary.ModelUID)
	if err != nil {
		return nil, err
	}

	return s.listActionsByBindings(
		ctx,
		bindings,
		nil,
		func(binding domain.PluginBinding, plugin domain.Plugin) (bool, error) {
			return s.resolver.bindingSatisfied(ctx, primary, binding)
		},
	)
}

func (s *service) listActionsForModel(
	ctx context.Context,
	bindings []domain.PluginBinding,
	pluginCache map[string]domain.Plugin,
) ([]pluginx.ResourceAction, error) {
	return s.listActionsByBindings(
		ctx,
		bindings,
		pluginCache,
		func(binding domain.PluginBinding, plugin domain.Plugin) (bool, error) {
			return true, nil
		},
	)
}

func (s *service) listActionsByBindings(
	ctx context.Context,
	bindings []domain.PluginBinding,
	pluginCache map[string]domain.Plugin,
	match bindingActionMatcher,
) ([]pluginx.ResourceAction, error) {
	if pluginCache == nil {
		pluginCache = map[string]domain.Plugin{}
	}

	actions := make([]pluginx.ResourceAction, 0, len(bindings))
	for _, binding := range bindings {
		plugin, err := s.loadCachedPlugin(ctx, binding.PluginID, pluginCache)
		if err != nil {
			return nil, err
		}

		ok, err := match(binding, plugin)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		actions = append(actions, pluginActions(plugin)...)
	}
	return actions, nil
}

func pluginActions(plugin domain.Plugin) []pluginx.ResourceAction {
	return lo.Map(plugin.Actions, func(action domain.PluginActionSpec, _ int) pluginx.ResourceAction {
		return toResourceAction(plugin, action)
	})
}

func (s *service) loadCachedPlugin(
	ctx context.Context,
	pluginID string,
	pluginCache map[string]domain.Plugin,
) (domain.Plugin, error) {
	if plugin, ok := pluginCache[pluginID]; ok {
		return plugin, nil
	}

	plugin, err := s.loadPlugin(ctx, pluginID)
	if err != nil {
		return domain.Plugin{}, err
	}
	pluginCache[pluginID] = plugin
	return plugin, nil
}
