package v192

import (
	"context"

	"github.com/Duke1616/ecmdb/cmd/initial/backup"
	"github.com/Duke1616/ecmdb/cmd/initial/incr"
	"github.com/Duke1616/ecmdb/cmd/initial/ioc"
	"github.com/gotomicro/ego/core/elog"
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



func (i *incrV192) Commit(ctx context.Context) error {
	i.logger.Info("开始执行 Commit", elog.String("版本", i.Version()))

	// 处理 casbin_rule 表
	if err := i.updateCasbinRule(ctx); err != nil {
		return err
	}

	i.logger.Info("Commit 执行完成")
	return nil
}

// updateCasbinRule 更新 casbin_rule 表
func (i *incrV192) updateCasbinRule(ctx context.Context) error {
	i.logger.Info("系统已不再依赖 MySQL，跳过 casbin_rule 表的增量数据迁移")
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
	i.logger.Info("系统已不再依赖 MySQL，跳过 casbin_rule 表的数据回滚")
	return nil
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

	// 跳过备份 MySQL casbin_rule 表
	i.logger.Info("系统已不再依赖 MySQL，跳过 casbin_rule 表的数据备份")

	// 备份 c_menu 集合
	_, err := backupManager.BackupMongoCollection(ctx, "c_menu", opts)
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
