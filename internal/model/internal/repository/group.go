package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/model/internal/domain"
	"github.com/Duke1616/ecmdb/internal/model/internal/repository/dao"
)

type MGRepository interface {
	CreateModelGroup(ctx context.Context, req domain.ModelGroup) (int64, error)
}

type groupRepository struct {
	dao dao.ModelGroupDAO
}

func NewMGRepository(dao dao.ModelGroupDAO) MGRepository {
	return &groupRepository{
		dao: dao,
	}
}

func (repo *groupRepository) CreateModelGroup(ctx context.Context, req domain.ModelGroup) (int64, error) {
	return repo.dao.CreateModelGroup(ctx, dao.ModelGroup{
		Name: req.Name,
	})
}
