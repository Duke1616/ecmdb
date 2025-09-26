package v192

import (
	"context"

	"github.com/Duke1616/ecmdb/cmd/initial/backup"
	"github.com/Duke1616/ecmdb/cmd/initial/incr"
	"github.com/Duke1616/ecmdb/cmd/initial/ioc"
	"github.com/Duke1616/ecmdb/cmd/initial/menu"
	"github.com/gotomicro/ego/core/elog"
	"gorm.io/gorm"
)

type incrV192 struct {
	App        *ioc.App
	ChangeSync *menu.ChangeSync
	logger     elog.Component
}

func NewIncrV192(app *ioc.App) incr.InitialIncr {
	return &incrV192{
		App:        app,
		ChangeSync: menu.NewChange(app),
		logger:     *elog.DefaultLogger,
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

	// 处理 casbin_rule 表
	if err := i.updateCasbinRule(ctx); err != nil {
		return err
	}

	// 处理菜单变更
	if err := i.ChangeSync.UpdateMenu(ctx); err != nil {
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
