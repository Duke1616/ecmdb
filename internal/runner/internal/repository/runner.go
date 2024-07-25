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
	FindByCodebookUid(ctx context.Context, codebookUid string, tag string) (domain.Runner, error)
	ListTagsPipelineByCodebookUid(ctx context.Context) ([]domain.RunnerTags, error)
}

func NewRunnerRepository(dao dao.RunnerDAO) RunnerRepository {
	return &runnerRepository{
		dao: dao,
	}
}

type runnerRepository struct {
	dao dao.RunnerDAO
}

func (repo *runnerRepository) ListTagsPipelineByCodebookUid(ctx context.Context) ([]domain.RunnerTags, error) {
	pipeline, err := repo.dao.ListTagsPipelineByCodebookUid(ctx)
	return slice.Map(pipeline, func(idx int, src dao.RunnerPipeline) domain.RunnerTags {
		var tags []string
		for _, runner := range src.RunnerTags {
			if runner.CodebookUid == src.CodebookUid {
				tags = append(tags, runner.Tags...)
			}
		}

		return domain.RunnerTags{
			CodebookUid: src.CodebookUid,
			Tags:        tags,
		}
	}), err
}

func (repo *runnerRepository) FindByCodebookUid(ctx context.Context, codebookUid string, tag string) (domain.Runner, error) {
	runner, err := repo.dao.FindByCodebookUid(ctx, codebookUid, tag)
	return repo.toDomain(runner), err
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
		CodebookSecret: req.CodebookSecret,
		CodebookUid:    req.CodebookUid,
		WorkerName:     req.WorkerName,
		Name:           req.Name,
		Tags:           req.Tags,
		Variables: slice.Map(req.Variables, func(idx int, src domain.Variables) dao.Variables {
			return dao.Variables{
				Key:   src.Key,
				Value: src.Value,
			}
		}),
		Desc:   req.Desc,
		Action: req.Action.ToUint8(),
	}
}

func (repo *runnerRepository) toDomain(req dao.Runner) domain.Runner {
	return domain.Runner{
		Id:             req.Id,
		Name:           req.Name,
		CodebookSecret: req.CodebookSecret,
		CodebookUid:    req.CodebookUid,
		WorkerName:     req.WorkerName,
		Tags:           req.Tags,
		Variables: slice.Map(req.Variables, func(idx int, src dao.Variables) domain.Variables {
			return domain.Variables{
				Key:   src.Key,
				Value: src.Value,
			}
		}),
		Desc:   req.Desc,
		Action: domain.Action(req.Action),
	}
}
