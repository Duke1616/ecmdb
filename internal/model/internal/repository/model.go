package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/model/internal/domain"
	"github.com/Duke1616/ecmdb/internal/model/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
	"time"
)

type ModelRepository interface {
	CreateModel(ctx context.Context, req domain.Model) (int64, error)
	FindModelById(ctx context.Context, id int64) (domain.Model, error)
	ListModels(ctx context.Context, offset, limit int64) ([]domain.Model, error)
	Total(ctx context.Context) (int64, error)
	ListModelByGroupIds(ctx context.Context, mgids []int64) ([]domain.Model, error)
	DeleteModelById(ctx context.Context, id int64) (int64, error)
	DeleteModelByUid(ctx context.Context, modelUid string) (int64, error)
}

func NewModelRepository(dao dao.ModelDAO) ModelRepository {
	return &modelRepository{
		dao: dao,
	}
}

type modelRepository struct {
	dao dao.ModelDAO
}

func (repo *modelRepository) CreateModel(ctx context.Context, req domain.Model) (int64, error) {
	return repo.dao.CreateModel(ctx, repo.toEntity(req))
}

func (repo *modelRepository) FindModelById(ctx context.Context, id int64) (domain.Model, error) {
	model, err := repo.dao.GetModelById(ctx, id)
	return repo.toDomain(model), err
}

func (repo *modelRepository) ListModels(ctx context.Context, offset, limit int64) ([]domain.Model, error) {
	models, err := repo.dao.ListModels(ctx, offset, limit)

	return slice.Map(models, func(idx int, src dao.Model) domain.Model {
		return repo.toDomain(src)
	}), err
}

func (repo *modelRepository) Total(ctx context.Context) (int64, error) {
	return repo.dao.Count(ctx)
}

func (repo *modelRepository) ListModelByGroupIds(ctx context.Context, mgids []int64) ([]domain.Model, error) {
	models, err := repo.dao.ListModelByGroupIds(ctx, mgids)

	return slice.Map(models, func(idx int, src dao.Model) domain.Model {
		return repo.toDomain(src)
	}), err
}

func (repo *modelRepository) DeleteModelById(ctx context.Context, id int64) (int64, error) {
	return repo.dao.DeleteModelById(ctx, id)
}

func (repo *modelRepository) DeleteModelByUid(ctx context.Context, modelUid string) (int64, error) {
	return repo.dao.DeleteModelByUid(ctx, modelUid)
}

func (repo *modelRepository) toEntity(req domain.Model) dao.Model {
	return dao.Model{
		ModelGroupId: req.GroupId,
		Name:         req.Name,
		UID:          req.UID,
		Icon:         req.Icon,
	}
}

func (repo *modelRepository) toDomain(modelDao dao.Model) domain.Model {
	return domain.Model{
		ID:      modelDao.Id,
		GroupId: modelDao.ModelGroupId,
		Name:    modelDao.Name,
		UID:     modelDao.UID,
		Icon:    modelDao.Icon,
		Ctime:   time.UnixMilli(modelDao.Ctime),
		Utime:   time.UnixMilli(modelDao.Utime),
	}
}
