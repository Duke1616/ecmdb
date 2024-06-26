package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/runner/internal/domain"
	"github.com/Duke1616/ecmdb/internal/runner/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type RunnerRepository interface {
	RegisterRunner(ctx context.Context, req domain.Runner) (int64, error)
	ListRunner(ctx context.Context, offset, limit int64) ([]domain.Runner, error)
	Total(ctx context.Context) (int64, error)
}

func NewRunnerRepository(dao dao.RunnerDAO) RunnerRepository {
	return &runnerRepository{
		dao: dao,
	}
}

type runnerRepository struct {
	dao dao.RunnerDAO
}

func (repo *runnerRepository) RegisterRunner(ctx context.Context, req domain.Runner) (int64, error) {
	return repo.dao.CreateRunner(ctx, repo.toEntity(req))
}

func (repo *runnerRepository) ListRunner(ctx context.Context, offset, limit int64) ([]domain.Runner, error) {
	ws, err := repo.dao.ListRunner(ctx, offset, limit)
	return slice.Map(ws, func(idx int, src dao.Runner) domain.Runner {
		return repo.toDomain(src)
	}), err
}

func (repo *runnerRepository) Total(ctx context.Context) (int64, error) {
	return repo.dao.Count(ctx)
}

func (repo *runnerRepository) toEntity(req domain.Runner) dao.Runner {
	return dao.Runner{
		TaskIdentifier: req.TaskIdentifier,
		TaskSecret:     req.TaskSecret,
		WorkName:       req.WorkName,
		Name:           req.Name,
		Tags:           req.Tags,
		Desc:           req.Desc,
		Action:         req.Action.ToUint8(),
	}
}

func (repo *runnerRepository) toDomain(req dao.Runner) domain.Runner {
	return domain.Runner{
		Id:             req.Id,
		Name:           req.Name,
		TaskIdentifier: req.TaskIdentifier,
		TaskSecret:     req.TaskSecret,
		WorkName:       req.WorkName,
		Tags:           req.Tags,
		Desc:           req.Desc,
		Action:         domain.Action(req.Action),
	}
}
