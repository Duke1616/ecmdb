package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/model/internal/domain"
	"github.com/Duke1616/ecmdb/internal/model/internal/repository/dao"
)

type ModelRepository interface {
	CreateModelGroup(ctx context.Context, req domain.ModelGroup) (int64, error)
}

func NewModelRepository(dao dao.ModelDAO) ModelRepository {
	return &modelRepository{
		dao: dao,
	}
}

type modelRepository struct {
	dao dao.ModelDAO
}

func (m *modelRepository) CreateModelGroup(ctx context.Context, req domain.ModelGroup) (int64, error) {
	return m.dao.CreateModelGroup(ctx, dao.ModelGroup{
		Name: req.Name,
	})
}
