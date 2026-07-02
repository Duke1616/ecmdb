package repository

import (
	"context"
	"time"

	"github.com/Duke1616/ecmdb/internal/domain"
	"github.com/Duke1616/ecmdb/internal/repository/dao"
	pluginx "github.com/Duke1616/ecmdb/pkg/plugin"
)

type PluginRepository interface {
	// UpsertPlugin 按插件 UID 创建或更新插件定义，并保留已有自增 ID。
	UpsertPlugin(ctx context.Context, p domain.Plugin) error

	// UpsertBinding 按绑定 UID 创建或更新插件和模型的输入映射关系。
	UpsertBinding(ctx context.Context, b domain.PluginBinding) error

	// UpdateBindingEnabled 更新绑定启停状态。
	UpdateBindingEnabled(ctx context.Context, uid string, enabled bool) error

	// GetPlugin 根据插件 UID 查询插件定义。
	GetPlugin(ctx context.Context, uid string) (domain.Plugin, error)

	// ListPlugins 查询插件定义列表。
	ListPlugins(ctx context.Context) ([]domain.Plugin, error)

	// ListBindingsByPluginID 查询插件绑定列表。
	ListBindingsByPluginID(ctx context.Context, pluginID string) ([]domain.PluginBinding, error)

	// ListBindingsByPluginIDs 批量查询插件绑定列表。
	ListBindingsByPluginIDs(ctx context.Context, pluginIDs []string) ([]domain.PluginBinding, error)

	// ListEnabledBindingsByModelUID 查询指定模型启用中的插件绑定。
	ListEnabledBindingsByModelUID(ctx context.Context, modelUID string) ([]domain.PluginBinding, error)

	// ListEnabledBindingsByModelUIDs 批量查询指定模型启用中的插件绑定。
	ListEnabledBindingsByModelUIDs(ctx context.Context, modelUIDs []string) ([]domain.PluginBinding, error)
}

type pluginRepository struct {
	dao dao.PluginDAO
}

func NewPluginRepository(dao dao.PluginDAO) PluginRepository {
	return &pluginRepository{dao: dao}
}

func (repo *pluginRepository) UpsertPlugin(ctx context.Context, p domain.Plugin) error {
	return repo.dao.UpsertPlugin(ctx, dao.Plugin{
		Id:      p.ID,
		UID:     p.UID,
		Name:    p.Name,
		Type:    p.Type,
		Version: p.Version,
		Actions: p.Actions,
	})
}

func (repo *pluginRepository) UpsertBinding(ctx context.Context, b domain.PluginBinding) error {
	return repo.dao.UpsertBinding(ctx, dao.PluginBinding{
		Id:       b.ID,
		UID:      b.UID,
		PluginID: b.PluginID,
		ModelUID: b.ModelUID,
		Enabled:  b.Enabled,
		Graph:    b.Graph,
	})
}

func (repo *pluginRepository) UpdateBindingEnabled(ctx context.Context, uid string, enabled bool) error {
	return repo.dao.UpdateBindingEnabled(ctx, uid, enabled)
}

func (repo *pluginRepository) GetPlugin(ctx context.Context, uid string) (domain.Plugin, error) {
	p, err := repo.dao.GetPlugin(ctx, uid)
	if err != nil {
		return domain.Plugin{}, err
	}
	return domain.Plugin{
		ID:      p.Id,
		UID:     p.UID,
		Name:    p.Name,
		Type:    p.Type,
		Version: p.Version,
		Actions: p.Actions,
		Ctime:   time.UnixMilli(p.Ctime).UnixMilli(),
		Utime:   time.UnixMilli(p.Utime).UnixMilli(),
	}, nil
}

func (repo *pluginRepository) ListPlugins(ctx context.Context) ([]domain.Plugin, error) {
	plugins, err := repo.dao.ListPlugins(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]domain.Plugin, 0, len(plugins))
	for _, plugin := range plugins {
		res = append(res, domain.Plugin{
			ID:      plugin.Id,
			UID:     plugin.UID,
			Name:    plugin.Name,
			Type:    plugin.Type,
			Version: plugin.Version,
			Actions: plugin.Actions,
			Ctime:   time.UnixMilli(plugin.Ctime).UnixMilli(),
			Utime:   time.UnixMilli(plugin.Utime).UnixMilli(),
		})
	}
	return res, nil
}

func (repo *pluginRepository) ListBindingsByPluginID(ctx context.Context, pluginID string) ([]domain.PluginBinding, error) {
	bindings, err := repo.dao.ListBindingsByPluginID(ctx, pluginID)
	if err != nil {
		return nil, err
	}
	return toPluginBindings(bindings), nil
}

func (repo *pluginRepository) ListBindingsByPluginIDs(ctx context.Context, pluginIDs []string) ([]domain.PluginBinding, error) {
	bindings, err := repo.dao.ListBindingsByPluginIDs(ctx, pluginIDs)
	if err != nil {
		return nil, err
	}
	return toPluginBindings(bindings), nil
}

func (repo *pluginRepository) ListEnabledBindingsByModelUID(ctx context.Context, modelUID string) ([]domain.PluginBinding, error) {
	bindings, err := repo.dao.ListEnabledBindingsByModelUID(ctx, modelUID)
	if err != nil {
		return nil, err
	}
	return toPluginBindings(bindings), nil
}

func (repo *pluginRepository) ListEnabledBindingsByModelUIDs(ctx context.Context, modelUIDs []string) ([]domain.PluginBinding, error) {
	bindings, err := repo.dao.ListEnabledBindingsByModelUIDs(ctx, modelUIDs)
	if err != nil {
		return nil, err
	}
	return toPluginBindings(bindings), nil
}

func toPluginBindings(bindings []dao.PluginBinding) []domain.PluginBinding {
	res := make([]domain.PluginBinding, 0, len(bindings))
	for _, binding := range bindings {
		res = append(res, domain.PluginBinding{
			ID:       binding.Id,
			UID:      binding.UID,
			PluginID: binding.PluginID,
			ModelUID: binding.ModelUID,
			Enabled:  binding.Enabled,
			Graph:    bindingGraph(binding),
			Ctime:    time.UnixMilli(binding.Ctime).UnixMilli(),
			Utime:    time.UnixMilli(binding.Utime).UnixMilli(),
		})
	}
	return res
}

func bindingGraph(binding dao.PluginBinding) *pluginx.BindingGraph {
	if binding.Graph != nil {
		return binding.Graph
	}
	if len(binding.LegacySpecs) == 0 {
		return nil
	}

	graph, err := pluginx.GraphFromBindingSpecs(binding.ModelUID, binding.LegacySpecs)
	if err != nil {
		return nil
	}
	return graph
}
