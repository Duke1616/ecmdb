package plugin

import (
	"context"
	"errors"

	"github.com/Duke1616/ecmdb/internal/errs"
	sshplugin "github.com/Duke1616/ecmdb/internal/plugin/ssh"
	pluginx "github.com/Duke1616/ecmdb/pkg/plugin"
)

func (s *service) SyncBuiltinDefinitions(ctx context.Context) error {
	for _, def := range builtinDefinitions() {
		if err := s.importDefinition(ctx, def); err != nil {
			return err
		}
	}
	return nil
}

func builtinDefinitions() []pluginx.Definition {
	return []pluginx.Definition{
		sshplugin.Definition(),
	}
}

func (s *service) syncBuiltinPlugin(ctx context.Context, next pluginx.Plugin) error {
	plugin, err := s.keepPluginRuntimeState(ctx, next)
	if err != nil {
		return err
	}
	return s.UpsertPlugin(ctx, plugin)
}

func (s *service) syncBuiltinBinding(ctx context.Context, next pluginx.Binding) error {
	binding, err := s.keepBindingRuntimeState(ctx, next)
	if err != nil {
		return err
	}
	return s.UpsertBinding(ctx, binding)
}

func (s *service) keepPluginRuntimeState(ctx context.Context, next pluginx.Plugin) (pluginx.Plugin, error) {
	existing, err := s.repo.GetPlugin(ctx, next.UID)
	switch {
	case err == nil:
		next.ID = existing.ID
		next.Enabled = existing.Enabled
		next.Config = existing.Config
		return next, nil
	case errors.Is(err, errs.ErrNotFound):
		return next, nil
	default:
		return pluginx.Plugin{}, err
	}
}

func (s *service) keepBindingRuntimeState(ctx context.Context, next pluginx.Binding) (pluginx.Binding, error) {
	existing, err := s.repo.GetBinding(ctx, next.UID)
	switch {
	case err == nil:
		next.ID = existing.ID
		next.Enabled = existing.Enabled
		next.Config = existing.Config
		return next, nil
	case errors.Is(err, errs.ErrNotFound):
		return next, nil
	default:
		return pluginx.Binding{}, err
	}
}
