package executor

import (
	"context"
	"errors"

	"github.com/Duke1616/ecmdb/internal/errs"
	"github.com/Duke1616/ecmdb/internal/model"
	"github.com/gotomicro/ego/core/elog"
)

// ModelExecutor 模型执行器，负责模型的创建
type ModelExecutor interface {
	// Execute 执行模型创建，直接使用 domain.Model
	Execute(ctx context.Context, m model.Model) (int64, error)
}

type modelExecutor struct {
	modelSvc model.Service
	logger   *elog.Component
}

// NewModelExecutor 创建模型执行器
func NewModelExecutor(modelSvc model.Service) ModelExecutor {
	return &modelExecutor{
		modelSvc: modelSvc,
		logger:   elog.DefaultLogger,
	}
}

// Execute 执行模型创建
func (e *modelExecutor) Execute(ctx context.Context, m model.Model) (int64, error) {
	e.logger.Info("开始创建模型", elog.String("模型UID", m.UID))

	// 检查模型是否已存在
	existingModel, err := e.modelSvc.GetByUid(ctx, m.UID)
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		e.logger.Error("检查模型失败",
			elog.String("模型UID", m.UID),
			elog.FieldErr(err))
		return 0, err
	}

	// 模型已存在，直接返回
	if existingModel.ID != 0 {
		e.logger.Info("模型已存在，跳过创建",
			elog.String("模型UID", m.UID),
			elog.Int64("模型ID", existingModel.ID))
		return existingModel.ID, nil
	}

	// 创建模型
	modelID, err := e.modelSvc.Create(ctx, m)
	if err != nil {
		e.logger.Error("创建模型失败",
			elog.String("模型UID", m.UID),
			elog.FieldErr(err))
		return 0, err
	}

	e.logger.Info("模型创建成功",
		elog.String("模型UID", m.UID),
		elog.Int64("模型ID", modelID))

	return modelID, nil
}
