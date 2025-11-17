package repository

import (
	"context"
	"time"

	"github.com/Duke1616/ecmdb/internal/model/internal/domain"
	"github.com/Duke1616/ecmdb/internal/model/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type ModelRepository interface {
	Create(ctx context.Context, req domain.Model) (int64, error)
	FindById(ctx context.Context, id int64) (domain.Model, error)
	GetByUids(ctx context.Context, uids []string) ([]domain.Model, error)
	List(ctx context.Context, offset, limit int64) ([]domain.Model, error)
	ListAll(ctx context.Context) ([]domain.Model, error)
	Total(ctx context.Context) (int64, error)
	ListByGroupIds(ctx context.Context, mgids []int64) ([]domain.Model, error)
	DeleteById(ctx context.Context, id int64) (int64, error)
	DeleteByUid(ctx context.Context, modelUid string) (int64, error)
}

func NewModelRepository(dao dao.ModelDAO) ModelRepository {
	return &modelRepository{
		dao: dao,
	}
}

type modelRepository struct {
	dao dao.ModelDAO
}

func (repo *modelRepository) GetByUids(ctx context.Context, uids []string) ([]domain.Model, error) {
	models, err := repo.dao.GetByUids(ctx, uids)

	return slice.Map(models, func(idx int, src dao.Model) domain.Model {
		return repo.toDomain(src)
	}), err
}

func (repo *modelRepository) ListAll(ctx context.Context) ([]domain.Model, error) {
	models, err := repo.dao.ListAll(ctx)

	return slice.Map(models, func(idx int, src dao.Model) domain.Model {
		return repo.toDomain(src)
	}), err
}

func (repo *modelRepository) Create(ctx context.Context, req domain.Model) (int64, error) {
	return repo.dao.Create(ctx, repo.toEntity(req))
}

func (repo *modelRepository) FindById(ctx context.Context, id int64) (domain.Model, error) {
	model, err := repo.dao.GetById(ctx, id)
	return repo.toDomain(model), err
}

func (repo *modelRepository) List(ctx context.Context, offset, limit int64) ([]domain.Model, error) {
	models, err := repo.dao.List(ctx, offset, limit)

	return slice.Map(models, func(idx int, src dao.Model) domain.Model {
		return repo.toDomain(src)
	}), err
}

func (repo *modelRepository) Total(ctx context.Context) (int64, error) {
	return repo.dao.Count(ctx)
}

func (repo *modelRepository) ListByGroupIds(ctx context.Context, mgids []int64) ([]domain.Model, error) {
	models, err := repo.dao.ListByGroupIds(ctx, mgids)

	return slice.Map(models, func(idx int, src dao.Model) domain.Model {
		return repo.toDomain(src)
	}), err
}

func (repo *modelRepository) DeleteById(ctx context.Context, id int64) (int64, error) {
	return repo.dao.DeleteById(ctx, id)
}

func (repo *modelRepository) DeleteByUid(ctx context.Context, modelUid string) (int64, error) {
	return repo.dao.DeleteByUid(ctx, modelUid)
}

func (repo *modelRepository) toEntity(req domain.Model) dao.Model {
	return dao.Model{
		ModelGroupId: req.GroupId,
		Builtin:      req.Builtin,
		Name:         req.Name,
		UID:          req.UID,
		Icon:         req.Icon,
	}
}

func (repo *modelRepository) toDomain(modelDao dao.Model) domain.Model {
	return domain.Model{
		ID:      modelDao.Id,
		GroupId: modelDao.ModelGroupId,
		Builtin: modelDao.Builtin,
		Name:    modelDao.Name,
		UID:     modelDao.UID,
		Icon:    modelDao.Icon,
		Ctime:   time.UnixMilli(modelDao.Ctime),
		Utime:   time.UnixMilli(modelDao.Utime),
	}
}
