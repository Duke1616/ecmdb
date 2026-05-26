package mongox

import (
	"context"
	"fmt"

	"github.com/Duke1616/eiam/pkg/ctxutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Plugin 代表全局 MongoDB 插件拦截器（IMongoPlugin 别名）
type Plugin IMongoPlugin

// DB 是支持插件生命周期管理与泛型数据集代理的 MongoDB 核心管理器
type DB struct {
	dbClient *mongo.Client
	dbName   string
	native   *mongo.Database
	plugins  []Plugin
}

// NewDB 实例化支持生命周期拦截的 DB 管理器
func NewDB(client *mongo.Client, dbName string) *DB {
	return &DB{
		dbClient: client,
		dbName:   dbName,
		native:   client.Database(dbName),
		plugins:  []Plugin{},
	}
}

// Use 注册插件拦截器（类似 GORM 的 db.Use()）
func (db *DB) Use(plugin Plugin) {
	db.plugins = append(db.plugins, plugin)
}

// Client 返回底层原生 *mongo.Client
func (db *DB) Client() *mongo.Client {
	return db.dbClient
}

// Database 返回默认原生 *mongo.Database
func (db *DB) Database() *mongo.Database {
	return db.native
}

// DatabaseWithTenant 物理隔离专用：根据 Context 动态路由并返回租户的物理数据库
func (db *DB) DatabaseWithTenant(ctx context.Context) *mongo.Database {
	tenantID := ctxutil.GetTenantID(ctx)
	if tenantID <= 0 {
		return db.native
	}
	dynamicDBName := fmt.Sprintf("%s_tenant_%d", db.dbName, tenantID.Int64())
	return db.dbClient.Database(dynamicDBName)
}

// CollectionWithTenant 物理隔离专用：获取租户物理隔离库的原生集合（不做逻辑拦截）
func (db *DB) CollectionWithTenant(ctx context.Context, collName string) *mongo.Collection {
	return db.DatabaseWithTenant(ctx).Collection(collName)
}

// ==========================================
// 插件生命周期核心分发总线
// ==========================================

func (db *DB) runBeforeFind(stmt *Statement) error {
	for _, p := range db.plugins {
		if hook, ok := p.(BeforeFinder); ok {
			if err := hook.BeforeFind(stmt); err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *DB) runBeforeInsert(stmt *Statement) error {
	for _, p := range db.plugins {
		if hook, ok := p.(BeforeInserter); ok {
			if err := hook.BeforeInsert(stmt); err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *DB) runBeforeUpdate(stmt *Statement) error {
	for _, p := range db.plugins {
		if hook, ok := p.(BeforeUpdater); ok {
			if err := hook.BeforeUpdate(stmt); err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *DB) runBeforeDelete(stmt *Statement) error {
	for _, p := range db.plugins {
		if hook, ok := p.(BeforeDeleter); ok {
			if err := hook.BeforeDelete(stmt); err != nil {
				return err
			}
		}
	}
	return nil
}

// GetBatchIdGenerator 从 c_id_generator 中原子地为指定集合申请 count 个连续自增 ID，供手动批量 Upsert 场景使用
func (db *DB) GetBatchIdGenerator(collection string, count int) (int64, error) {
	coll := db.native.Collection("c_id_generator")
	var result struct {
		Name   string `bson:"name"`
		NextID int64  `bson:"next_id"`
	}

	upsert := true
	returnAfter := options.After
	err := coll.FindOneAndUpdate(
		context.Background(),
		bson.M{"name": collection},
		bson.M{"$inc": bson.M{"next_id": int64(count)}},
		&options.FindOneAndUpdateOptions{
			Upsert:         &upsert,
			ReturnDocument: &returnAfter,
		},
	).Decode(&result)

	if err != nil {
		return 0, err
	}

	return result.NextID - int64(count) + 1, nil
}
