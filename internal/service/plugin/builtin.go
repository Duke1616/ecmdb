package plugin

import (
	"context"
	"errors"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/errs"
	sshplugin "github.com/Duke1616/ecmdb/internal/plugin/ssh"
	pluginx "github.com/Duke1616/ecmdb/pkg/plugin"
)

func (s *service) RegisterBuiltinPlugins(ctx context.Context) error {
	for _, def := range builtinDefinitions() {
		if err := s.syncBuiltinPlugin(ctx, def.Plugin); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) SyncDefaultSchema(ctx context.Context, pluginID string) error {
	var targetDef *pluginx.Definition
	for _, def := range builtinDefinitions() {
		if def.Plugin.UID == pluginID {
			targetDef = &def
			break
		}
	}
	if targetDef == nil {
		return fmt.Errorf("未找到对应的内置插件定义: %s", pluginID)
	}

	// 按需导入内置的模型、属性组、以及模型关联类型关系
	if err := s.importSchema(ctx, targetDef.Schema); err != nil {
		return err
	}

	// 同步生成对应的内置默认 Binding 记录
	for _, binding := range targetDef.Bindings {
		if err := s.syncBuiltinBinding(ctx, binding); err != nil {
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
		return next, nil
	case errors.Is(err, errs.ErrNotFound):
		return next, nil
	default:
		return pluginx.Binding{}, err
	}
}
