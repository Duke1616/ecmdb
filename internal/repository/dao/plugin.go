package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/internal/errs"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/Duke1616/ecmdb/pkg/plugin"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	PluginCollection        = "c_plugins"
	PluginBindingCollection = "c_plugin_bindings"
)

type Plugin struct {
	TenantID int64               `bson:"tenant_id,omitempty"`
	Id       int64               `bson:"id"`
	UID      string              `bson:"uid"`
	Name     string              `bson:"name"`
	Type     string              `bson:"type"`
	Version  string              `bson:"version"`
	Enabled  bool                `bson:"enabled"`
	Actions  []plugin.ActionSpec `bson:"actions"`
	Config   map[string]any      `bson:"config,omitempty"`
	Ctime    int64               `bson:"ctime"`
	Utime    int64               `bson:"utime"`
}

type PluginBinding struct {
	TenantID int64                 `bson:"tenant_id,omitempty"`
	Id       int64                 `bson:"id"`
	UID      string                `bson:"uid"`
	PluginID string                `bson:"plugin_id"`
	ModelUID string                `bson:"model_uid"`
	Enabled  bool                  `bson:"enabled"`
	Specs    []plugin.ResourceSpec `bson:"specs"`
	Config   map[string]any        `bson:"config,omitempty"`
	Ctime    int64                 `bson:"ctime"`
	Utime    int64                 `bson:"utime"`
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

	// GetPlugin 根据 UID 查询插件存储记录。
	GetPlugin(ctx context.Context, uid string) (Plugin, error)

	// GetBinding 根据 UID 查询插件绑定存储记录。
	GetBinding(ctx context.Context, uid string) (PluginBinding, error)

	// ListEnabledBindingsByModelUID 查询指定模型启用中的插件绑定记录。
	ListEnabledBindingsByModelUID(ctx context.Context, modelUID string) ([]PluginBinding, error)
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
				"enabled": p.Enabled,
				"actions": p.Actions,
				"config":  p.Config,
				"utime":   p.Utime,
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
				"specs":     b.Specs,
				"config":    b.Config,
				"utime":     b.Utime,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("upsert plugin binding: %w", err)
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

func (dao *pluginDAO) GetBinding(ctx context.Context, uid string) (PluginBinding, error) {
	b, err := dao.bindingColl.FindOne(ctx, bson.M{"uid": uid})
	if err != nil {
		if mongox.IsNotFoundError(err) {
			return PluginBinding{}, fmt.Errorf("插件绑定查询: %w", errs.ErrNotFound)
		}
		return PluginBinding{}, fmt.Errorf("插件绑定查询失败: %w", err)
	}
	return *b, nil
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
