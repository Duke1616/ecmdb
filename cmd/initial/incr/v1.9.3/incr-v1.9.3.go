package v193

import (
	"context"

	"github.com/Duke1616/ecmdb/cmd/initial/incr"
	"github.com/Duke1616/ecmdb/cmd/initial/ioc"
	"github.com/gotomicro/ego/core/elog"
)

type incrV193 struct {
	App    *ioc.App
	logger elog.Component
}

func NewIncrV193(app *ioc.App) incr.InitialIncr {
	return &incrV193{
		App:    app,
		logger: *elog.DefaultLogger,
	}
}

func (i *incrV193) Version() string {
	return "v1.9.3"
}

func (i *incrV193) Commit(ctx context.Context) error {
	i.logger.Info("开始执行 Commit（历史版本已归档）", elog.String("版本", i.Version()))
	return nil
}

func (i *incrV193) Rollback(ctx context.Context) error {
	return nil
}

func (i *incrV193) Before(ctx context.Context) error {
	return nil
}

func (i *incrV193) After(ctx context.Context) error {
	i.logger.Info("开始执行 After，更新版本信息", elog.String("版本", i.Version()))
	if err := i.App.VerSvc.CreateOrUpdateVersion(ctx, i.Version()); err != nil {
		i.logger.Error("更新版本信息失败", elog.FieldErr(err))
		return err
	}
	i.logger.Info("After 执行完成，版本信息已更新")
	return nil
}
