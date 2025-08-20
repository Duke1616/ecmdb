package repository

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/model/internal/domain"
	"github.com/Duke1616/ecmdb/internal/model/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type MGRepository interface {
	CreateModelGroup(ctx context.Context, req domain.ModelGroup) (int64, error)
	List(ctx context.Context, offset, limit int64) ([]domain.ModelGroup, error)
	Total(ctx context.Context) (int64, error)
	DeleteModelGroup(ctx context.Context, id int64) (int64, error)
}

type groupRepository struct {
	dao dao.ModelGroupDAO
}

func NewMGRepository(dao dao.ModelGroupDAO) MGRepository {
	return &groupRepository{
		dao: dao,
	}
}

func (repo *groupRepository) List(ctx context.Context, offset, limit int64) ([]domain.ModelGroup, error) {
	mgs, err := repo.dao.List(ctx, offset, limit)
	return slice.Map(mgs, func(idx int, src dao.ModelGroup) domain.ModelGroup {
		return repo.toDomain(src)
	}), err
}

func (repo *groupRepository) Total(ctx context.Context) (int64, error) {
	return repo.dao.Count(ctx)
}

func (repo *groupRepository) CreateModelGroup(ctx context.Context, req domain.ModelGroup) (int64, error) {
	return repo.dao.CreateModelGroup(ctx, dao.ModelGroup{
		Name: req.Name,
	})
}

func (repo *groupRepository) DeleteModelGroup(ctx context.Context, id int64) (int64, error) {
	return repo.dao.Delete(ctx, id)
}

func (repo *groupRepository) toDomain(modelDao dao.ModelGroup) domain.ModelGroup {
	return domain.ModelGroup{
		ID:   modelDao.Id,
		Name: modelDao.Name,
	}
}
