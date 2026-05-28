package v195

import (
	"context"
	"fmt"

	"github.com/Duke1616/ecmdb/cmd/initial/backup"
	"github.com/Duke1616/ecmdb/cmd/initial/incr"
	"github.com/Duke1616/ecmdb/cmd/initial/ioc"
	"github.com/gotomicro/ego/core/elog"
	"go.mongodb.org/mongo-driver/bson"
)

const defaultTargetTenantID int64 = 2

// NOTE: 将 CMDB 8 大核心集合名称统一抽取为包级常量，消灭硬编码冗余 (DRY 原则)
var cmdbCollections = []string{
	"c_resources",
	"c_model",
	"c_model_group",
	"c_attribute",
	"c_attribute_group",
	"c_relation_type",
	"c_relation_model",
	"c_relation_resource",
}

type incrV195 struct {
	App       *ioc.App
	logger    elog.Component
	backupIDs map[string]string // 存储 Before 阶段每个集合备份成功对应的 BackupID，用于精确原子回滚自愈
}

func NewIncrV195(app *ioc.App) incr.InitialIncr {
	return &incrV195{
		App:       app,
		logger:    *elog.DefaultLogger,
		backupIDs: make(map[string]string),
	}
}

func (i *incrV195) Version() string {
	return "v1.9.5"
}

func (i *incrV195) Commit(ctx context.Context) error {
	i.logger.Info("开始执行 Commit 批量迁移历史无租户数据", elog.String("版本", i.Version()))
	for _, collName := range cmdbCollections {
		col := i.App.DB.Collection(collName)

		// 批量把没有 tenant_id 或 tenant_id == 0 的历史数据迁移至指定的租户空间
		result, err := col.UpdateMany(ctx,
			bson.M{"$or": []bson.M{
				{"tenant_id": bson.M{"$exists": false}},
				{"tenant_id": 0},
			}},
			bson.M{"$set": bson.M{"tenant_id": defaultTargetTenantID}},
		)
		if err != nil {
			return fmt.Errorf("清洗集合 %s 失败: %w", collName, err)
		}

		i.logger.Info("成功迁移清洗集合数据",
			elog.String("collection", collName),
			elog.Int64("matched", result.MatchedCount),
			elog.Int64("modified", result.ModifiedCount),
		)
	}

	i.logger.Info("Commit 执行完成", elog.String("版本", i.Version()))
	return nil
}

func (i *incrV195) Rollback(ctx context.Context) error {
	i.logger.Info("升级发生异常，开始执行原子级 Rollback 灾备自愈还原...", elog.String("版本", i.Version()))
	backupManager := backup.NewBackupManager(i.App)

	// 精准、原子性自愈。根据 Before 阶段实际备份成功的 BackupID，精准倒序还原每一个被修改过的集合
	for j := len(cmdbCollections) - 1; j >= 0; j-- {
		collName := cmdbCollections[j]
		backupID, ok := i.backupIDs[collName]
		if !ok {
			i.logger.Warn("集合未找到对应的成功备份日志，跳过恢复", elog.String("collection", collName))
			continue
		}

		i.logger.Info("正在恢复集合到迁移前状态...", elog.String("collection", collName), elog.String("backupID", backupID))
		if err := backupManager.RestoreMongoCollection(ctx, collName, backupID); err != nil {
			i.logger.Error("回滚灾备还原失败", elog.String("collection", collName), elog.FieldErr(err))
			return fmt.Errorf("回滚灾备还原失败: %w", err)
		}
	}

	i.logger.Info("Rollback 执行完成，系统数据完美自愈复原", elog.String("版本", i.Version()))
	return nil
}

func (i *incrV195) Before(ctx context.Context) error {
	i.logger.Info("开始执行 Before，备份 CMDB 数据", elog.String("版本", i.Version()))

	backupManager := backup.NewBackupManager(i.App)

	for _, collName := range cmdbCollections {
		opts := backup.Options{
			Version:     i.Version(),
			Description: fmt.Sprintf("%s 版本升级前备份集合 %s", i.Version(), collName),
			Tags: map[string]string{
				"type":   "version_upgrade",
				"module": "cmdb_migration",
			},
		}

		// 强阻断机制。只要任何一个集合备份失败，立刻抛出错误强阻断升级，杜绝裸奔 Commit 清洗
		res, err := backupManager.BackupMongoCollection(ctx, collName, opts)
		if err != nil {
			return fmt.Errorf("升级前置备份 %s 失败，已触发强阻断终止升级: %w", collName, err)
		}

		// 妥善缓存备份批次 ID，以备 Rollback 阶段使用
		i.backupIDs[collName] = res.BackupID
		i.logger.Info("备份集合成功", elog.String("collection", collName), elog.String("backupID", res.BackupID))
	}

	i.logger.Info("Before 执行完成，CMDB 备份完成")
	return nil
}

func (i *incrV195) After(ctx context.Context) error {
	i.logger.Info("开始执行 After，更新版本信息", elog.String("版本", i.Version()))
	if err := i.App.VerSvc.CreateOrUpdateVersion(ctx, i.Version()); err != nil {
		i.logger.Error("更新版本信息失败", elog.FieldErr(err))
		return err
	}
	i.logger.Info("After 执行完成，版本信息已更新")
	return nil
}
