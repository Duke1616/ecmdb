package dao

import (
	"context"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"gorm.io/gorm"
)

type ProcessEngineDAO interface {
	TodoCount(ctx context.Context, userId, processName string) (int64, error)
}

type processEngineDAO struct {
	db *gorm.DB
}

func NewProcessEngineDAO(db *gorm.DB) ProcessEngineDAO {
	return &processEngineDAO{
		db: db,
	}
}

func (g *processEngineDAO) TodoCount(ctx context.Context, userId, processName string) (int64, error) {
	var res int64
	db := g.db.WithContext(ctx).Model(&model.Task{}).Table("proc_task")
	// 根据 userId 是否为空添加条件
	if userId != "" {
		db = db.Where("user_id = ?", userId)
	}
	if processName != "" {
		db = db.Where("process_name = ?", processName)
	}

	db = db.Where("is_finished = ?", 0)
	err := db.Count(&res).Error
	return res, err
}
