package v192

import (
	"context"
	"time"

	"github.com/Duke1616/ecmdb/cmd/initial/backup"
	"github.com/Duke1616/ecmdb/cmd/initial/incr"
	"github.com/Duke1616/ecmdb/cmd/initial/ioc"
	"github.com/gotomicro/ego/core/elog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type incrV192 struct {
	App    *ioc.App
	logger elog.Component
}

func NewIncrV192(app *ioc.App) incr.InitialIncr {
	return &incrV192{
		App:    app,
		logger: *elog.DefaultLogger,
	}
}

func (i *incrV192) Version() string {
	return "v1.9.2"
}

// 执行更新语句的小工具函数
func (i *incrV192) update(tx *gorm.DB, column string, value interface{}) error {
	return tx.Table("casbin_rule").
		Where("ptype = ?", "p").
		Update(column, value).Error
}

func (i *incrV192) Commit(ctx context.Context) error {
	i.logger.Info("开始执行 Commit", elog.String("版本", i.Version()))

	// 处理 c_menu 集合中的 endpoints
	if err := i.updateMenuEndpoints(ctx); err != nil {
		return err
	}

	// 处理 casbin_rule 表
	if err := i.updateCasbinRule(ctx); err != nil {
		return err
	}

	i.logger.Info("Commit 执行完成")
	return nil
}

// updateCasbinRule 更新 casbin_rule 表
func (i *incrV192) updateCasbinRule(ctx context.Context) error {
	return i.App.GormDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Table("casbin_rule").
			Where("ptype = ? AND v4 IS NOT NULL AND v4 != ''", "p").
			Count(&count).Error; err != nil {
			i.logger.Error("检查 v4 是否已有数据失败", elog.FieldErr(err))
			return err
		}
		if count > 0 {
			i.logger.Info("v4 已经有数据，casbin_rule 更新跳过")
			return nil
		}

		if err := i.update(tx, "v4", gorm.Expr("v3")); err != nil {
			i.logger.Error("v3 -> v4 迁移失败", elog.FieldErr(err))
			return err
		}
		i.logger.Info("v3 -> v4 迁移完成")

		if err := i.update(tx, "v3", "CMDB"); err != nil {
			i.logger.Error("v3 更新为 'CMDB' 失败", elog.FieldErr(err))
			return err
		}
		i.logger.Info("v3 已更新为 'CMDB'")

		return nil
	})
}

// updateMenuEndpoints 更新 c_menu 集合中的 endpoints resource 字段
func (i *incrV192) updateMenuEndpoints(ctx context.Context) error {
	i.logger.Info("开始处理 c_menu 集合中的 endpoints resource 字段")

	collection := i.App.DB.Collection("c_menu")

	// 查询所有菜单数据
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		i.logger.Error("查询 c_menu 集合失败", elog.FieldErr(err))
		return err
	}
	defer cursor.Close(ctx)

	// 使用 cursor.All() 获取所有数据
	var menus []bson.M
	if err = cursor.All(ctx, &menus); err != nil {
		i.logger.Error("解码菜单数据失败", elog.FieldErr(err))
		return err
	}

	// 准备批量更新操作
	var bulkOperations []mongo.WriteModel
	var updatedCount int64

	for _, menu := range menus {
		// 检查是否有 endpoints 字段
		endpoints, ok := menu["endpoints"].(bson.A)
		if !ok || len(endpoints) == 0 {
			continue
		}

		// 检查是否需要更新
		needUpdate := false
		var updatedEndpoints []bson.M
		for _, endpoint := range endpoints {
			if ep, ok := endpoint.(bson.M); ok {
				resource, _ := ep["resource"].(string)
				if resource == "" {
					needUpdate = true
					ep["resource"] = "CMDB"
					i.logger.Info("准备更新 endpoint resource 字段",
						elog.String("path", ep["path"].(string)),
						elog.String("method", ep["method"].(string)))
				}
				updatedEndpoints = append(updatedEndpoints, ep)
			}
		}

		if !needUpdate {
			continue
		}

		// 添加到批量更新操作
		filter := bson.M{"id": menu["id"]}
		update := bson.M{
			"$set": bson.M{
				"endpoints": updatedEndpoints,
				"utime":     time.Now().UnixMilli(),
			},
		}

		bulkOperations = append(bulkOperations, mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update))
	}

	// 执行批量更新
	if len(bulkOperations) > 0 {
		i.logger.Info("开始执行批量更新", elog.Int("操作数量", len(bulkOperations)))

		// 分批执行，每批最多 1000 个操作
		batchSize := 1000
		for idx := 0; idx < len(bulkOperations); idx += batchSize {
			end := idx + batchSize
			if end > len(bulkOperations) {
				end = len(bulkOperations)
			}

			batch := bulkOperations[idx:end]
			result, err := collection.BulkWrite(ctx, batch)
			if err != nil {
				i.logger.Error("批量更新失败",
					elog.FieldErr(err),
					elog.Int("批次", idx/batchSize+1))
				return err
			}

			updatedCount += result.ModifiedCount
			i.logger.Info("批次更新完成",
				elog.Int("批次", idx/batchSize+1),
				elog.Int64("更新数量", result.ModifiedCount))
		}
	}

	i.logger.Info("c_menu 集合 endpoints 处理完成", elog.Int64("总更新数量", updatedCount))
	return nil
}

func (i *incrV192) Rollback(ctx context.Context) error {
	i.logger.Info("开始执行 Rollback", elog.String("版本", i.Version()))

	// 回滚 casbin_rule 表
	if err := i.rollbackCasbinRule(ctx); err != nil {
		return err
	}

	i.logger.Info("Rollback 执行完成")
	return nil
}

// rollbackCasbinRule 回滚 casbin_rule 表
func (i *incrV192) rollbackCasbinRule(ctx context.Context) error {
	return i.App.GormDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Table("casbin_rule").
			Where("ptype = ? AND v4 IS NOT NULL AND v4 != ''", "p").
			Count(&count).Error; err != nil {
			i.logger.Error("检查 v4 是否有数据失败", elog.FieldErr(err))
			return err
		}
		if count == 0 {
			i.logger.Info("v4 为空，casbin_rule 回滚跳过")
			return nil
		}

		if err := i.update(tx, "v3", gorm.Expr("v4")); err != nil {
			i.logger.Error("v4 -> v3 回滚失败", elog.FieldErr(err))
			return err
		}
		i.logger.Info("v4 -> v3 回滚完成")

		if err := i.update(tx, "v4", ""); err != nil {
			i.logger.Error("清空 v4 失败", elog.FieldErr(err))
			return err
		}
		i.logger.Info("v4 已清空")

		return nil
	})
}

func (i *incrV192) Before(ctx context.Context) error {
	i.logger.Info("开始执行 Before，备份数据", elog.String("版本", i.Version()))

	// 创建备份管理器
	backupManager := backup.NewBackupManager(i.App)

	// 备份选项
	opts := backup.Options{
		Version:     i.Version(),
		Description: "v1.9.2 版本更新前备份",
		Tags: map[string]string{
			"type":   "version_upgrade",
			"module": "menu_endpoints",
		},
	}

	// 备份 casbin_rule 表
	_, err := backupManager.BackupMySQLTable(ctx, "casbin_rule", opts)
	if err != nil {
		i.logger.Error("备份 casbin_rule 表失败", elog.FieldErr(err))
		return err
	}

	// 备份 c_menu 集合
	_, err = backupManager.BackupMongoCollection(ctx, "c_menu", opts)
	if err != nil {
		i.logger.Error("备份 c_menu 集合失败", elog.FieldErr(err))
		return err
	}

	i.logger.Info("Before 执行完成，数据备份完成")
	return nil
}

func (i *incrV192) After(ctx context.Context) error {
	i.logger.Info("开始执行 After，更新版本信息", elog.String("版本", i.Version()))
	if err := i.App.VerSvc.CreateOrUpdateVersion(ctx, i.Version()); err != nil {
		i.logger.Error("更新版本信息失败", elog.FieldErr(err))
		return err
	}
	i.logger.Info("After 执行完成，版本信息已更新")
	return nil
}
