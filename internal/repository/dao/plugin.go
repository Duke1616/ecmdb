package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/internal/errs"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/Duke1616/ecmdb/pkg/plugin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	PluginCollection        = "c_plugins"
	PluginBindingCollection = "c_plugin_bindings"
)

type Plugin struct {
	TenantID int64               `bson:"tenant_id" eiam:"shared:type=builtin"`
	Id       int64               `bson:"id"`
	UID      string              `bson:"uid"`
	Name     string              `bson:"name"`
	Type     string              `bson:"type"`
	Version  string              `bson:"version"`
	Actions  []plugin.ActionSpec `bson:"actions"`
	Ctime    int64               `bson:"ctime"`
	Utime    int64               `bson:"utime"`
}

type PluginBinding struct {
	TenantID    int64                 `bson:"tenant_id" eiam:"private"`
	Id          int64                 `bson:"id"`
	UID         string                `bson:"uid"`
	PluginID    string                `bson:"plugin_id"`
	ModelUID    string                `bson:"model_uid"`
	Enabled     bool                  `bson:"enabled"`
	Graph       *plugin.BindingGraph  `bson:"graph,omitempty"`
	LegacySpecs []plugin.ResourceSpec `bson:"specs,omitempty"`
	Ctime       int64                 `bson:"ctime"`
	Utime       int64                 `bson:"utime"`
}

func (p *Plugin) SetID(id int64) {
	p.Id = id
}

func (p *Plugin) GetID() int64 {
	return p.Id
}

func (b *PluginBinding) SetID(id int64) {
	b.Id = id
}

func (b *PluginBinding) GetID() int64 {
	return b.Id
}

type PluginDAO interface {
	// UpsertPlugin 按 UID 创建或更新插件存储记录。
	UpsertPlugin(ctx context.Context, p Plugin) error

	// UpsertBinding 按 UID 创建或更新插件绑定存储记录。
	UpsertBinding(ctx context.Context, b PluginBinding) error

	// UpdateBindingEnabled 更新指定绑定的启停状态。
	UpdateBindingEnabled(ctx context.Context, uid string, enabled bool) error

	// GetPlugin 根据 UID 查询插件存储记录。
	GetPlugin(ctx context.Context, uid string) (Plugin, error)

	// ListPlugins 查询全部插件记录。
	ListPlugins(ctx context.Context) ([]Plugin, error)

	// ListBindingsByPluginID 查询指定插件的绑定记录。
	ListBindingsByPluginID(ctx context.Context, pluginID string) ([]PluginBinding, error)

	// ListBindingsByPluginIDs 批量查询指定插件的绑定记录。
	ListBindingsByPluginIDs(ctx context.Context, pluginIDs []string) ([]PluginBinding, error)

	// ListEnabledBindingsByModelUID 查询指定模型启用中的插件绑定记录。
	ListEnabledBindingsByModelUID(ctx context.Context, modelUID string) ([]PluginBinding, error)

	// ListEnabledBindingsByModelUIDs 批量查询指定模型启用中的插件绑定记录。
	ListEnabledBindingsByModelUIDs(ctx context.Context, modelUIDs []string) ([]PluginBinding, error)
}

type pluginDAO struct {
	pluginColl  *mongox.Collection[Plugin]
	bindingColl *mongox.Collection[PluginBinding]
}

func NewPluginDAO(db *mongox.DB) PluginDAO {
	return &pluginDAO{
		pluginColl:  mongox.NewCollection[Plugin](db, PluginCollection),
		bindingColl: mongox.NewCollection[PluginBinding](db, PluginBindingCollection),
	}
}

func (dao *pluginDAO) UpsertPlugin(ctx context.Context, p Plugin) error {
	now := time.Now().UnixMilli()
	p.Utime = now
	if p.Ctime == 0 {
		p.Ctime = now
	}

	existing, err := dao.pluginColl.FindOne(ctx, bson.M{"uid": p.UID})
	if err != nil {
		if !mongox.IsNotFoundError(err) {
			return fmt.Errorf("查询插件失败: %w", err)
		}
		_, err = dao.pluginColl.InsertOne(ctx, &p)
		if err != nil {
			return fmt.Errorf("创建插件失败: %w", err)
		}
		return nil
	}
	p.Id = existing.Id

	_, err = dao.pluginColl.UpdateOne(ctx,
		bson.M{"uid": p.UID},
		bson.M{
			"$set": bson.M{
				"id":      p.Id,
				"name":    p.Name,
				"type":    p.Type,
				"version": p.Version,
				"actions": p.Actions,
				"utime":   p.Utime,
			},
			"$unset": bson.M{
				"input_specs": "",
			},
		},
	)
	if err != nil {
		return fmt.Errorf("upsert plugin: %w", err)
	}
	return nil
}

func (dao *pluginDAO) UpsertBinding(ctx context.Context, b PluginBinding) error {
	now := time.Now().UnixMilli()
	b.Utime = now
	if b.Ctime == 0 {
		b.Ctime = now
	}

	existing, err := dao.bindingColl.FindOne(ctx, bson.M{"uid": b.UID})
	if err != nil {
		if !mongox.IsNotFoundError(err) {
			return fmt.Errorf("查询插件绑定失败: %w", err)
		}
		_, err = dao.bindingColl.InsertOne(ctx, &b)
		if err != nil {
			return fmt.Errorf("创建插件绑定失败: %w", err)
		}
		return nil
	}
	b.Id = existing.Id

	_, err = dao.bindingColl.UpdateOne(ctx,
		bson.M{"uid": b.UID},
		bson.M{
			"$set": bson.M{
				"id":        b.Id,
				"plugin_id": b.PluginID,
				"model_uid": b.ModelUID,
				"enabled":   b.Enabled,
				"graph":     b.Graph,
				"utime":     b.Utime,
			},
			"$unset": bson.M{
				"specs": "",
			},
		},
	)
	if err != nil {
		return fmt.Errorf("upsert plugin binding: %w", err)
	}
	return nil
}

func (dao *pluginDAO) UpdateBindingEnabled(ctx context.Context, uid string, enabled bool) error {
	res, err := dao.bindingColl.UpdateOne(ctx,
		bson.M{"uid": uid},
		bson.M{
			"$set": bson.M{
				"enabled": enabled,
				"utime":   time.Now().UnixMilli(),
			},
		},
	)
	if err != nil {
		return fmt.Errorf("更新插件绑定启停状态失败: %w", err)
	}
	if res.MatchedCount == 0 {
		return fmt.Errorf("插件绑定查询: %w", errs.ErrNotFound)
	}
	return nil
}

func (dao *pluginDAO) GetPlugin(ctx context.Context, uid string) (Plugin, error) {
	p, err := dao.pluginColl.FindOne(ctx, bson.M{"uid": uid})
	if err != nil {
		if mongox.IsNotFoundError(err) {
			return Plugin{}, fmt.Errorf("插件查询: %w", errs.ErrNotFound)
		}
		return Plugin{}, fmt.Errorf("插件查询失败: %w", err)
	}
	return *p, nil
}

func (dao *pluginDAO) ListPlugins(ctx context.Context) ([]Plugin, error) {
	opts := &options.FindOptions{
		Sort: bson.D{{Key: "utime", Value: -1}},
	}

	plugins, err := dao.pluginColl.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("插件查询失败: %w", err)
	}
	return plugins, nil
}

func (dao *pluginDAO) ListBindingsByPluginID(ctx context.Context, pluginID string) ([]PluginBinding, error) {
	opts := &options.FindOptions{
		Sort: bson.D{{Key: "utime", Value: -1}},
	}

	bindings, err := dao.bindingColl.Find(ctx, bson.M{"plugin_id": pluginID}, opts)
	if err != nil {
		return nil, fmt.Errorf("插件绑定查询失败: %w", err)
	}
	return bindings, nil
}

func (dao *pluginDAO) ListBindingsByPluginIDs(ctx context.Context, pluginIDs []string) ([]PluginBinding, error) {
	if len(pluginIDs) == 0 {
		return []PluginBinding{}, nil
	}

	opts := &options.FindOptions{
		Sort: bson.D{{Key: "utime", Value: -1}},
	}

	bindings, err := dao.bindingColl.Find(ctx, bson.M{
		"plugin_id": bson.M{"$in": pluginIDs},
	}, opts)
	if err != nil {
		return nil, fmt.Errorf("插件绑定查询失败: %w", err)
	}
	return bindings, nil
}

func (dao *pluginDAO) ListEnabledBindingsByModelUID(ctx context.Context, modelUID string) ([]PluginBinding, error) {
	bindings, err := dao.bindingColl.Find(ctx, bson.M{
		"model_uid": modelUID,
		"enabled":   true,
	})
	if err != nil {
		return nil, fmt.Errorf("插件绑定查询失败: %w", err)
	}
	return bindings, nil
}

func (dao *pluginDAO) ListEnabledBindingsByModelUIDs(ctx context.Context, modelUIDs []string) ([]PluginBinding, error) {
	bindings, err := dao.bindingColl.Find(ctx, bson.M{
		"model_uid": bson.M{"$in": modelUIDs},
		"enabled":   true,
	})
	if err != nil {
		return nil, fmt.Errorf("插件绑定查询失败: %w", err)
	}
	return bindings, nil
}
