package repository

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/worker/internal/domain"
	"github.com/Duke1616/ecmdb/internal/worker/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type WorkerRepository interface {
	CreateWorker(ctx context.Context, req domain.Worker) (int64, error)
	FindByName(ctx context.Context, name string) (domain.Worker, error)
	FindByKey(ctx context.Context, key string) (domain.Worker, error)
	ListWorker(ctx context.Context, offset, limit int64) ([]domain.Worker, error)
	UpdateStatus(ctx context.Context, id int64, status uint8) (int64, error)
	Total(ctx context.Context) (int64, error)
}

func NewWorkerRepository(dao dao.WorkerDAO) WorkerRepository {
	return &workerRepository{
		dao: dao,
	}
}

type workerRepository struct {
	dao dao.WorkerDAO
}

func (repo *workerRepository) CreateWorker(ctx context.Context, req domain.Worker) (int64, error) {
	return repo.dao.CreateWorker(ctx, repo.toEntity(req))
}

func (repo *workerRepository) FindByName(ctx context.Context, name string) (domain.Worker, error) {
	worker, err := repo.dao.FindByName(ctx, name)
	return repo.toDomain(worker), err
}

func (repo *workerRepository) FindByKey(ctx context.Context, key string) (domain.Worker, error) {
	worker, err := repo.dao.FindByKey(ctx, key)
	return repo.toDomain(worker), err
}

func (repo *workerRepository) UpdateStatus(ctx context.Context, id int64, status uint8) (int64, error) {
	return repo.dao.UpdateStatus(ctx, id, status)
}

func (repo *workerRepository) ListWorker(ctx context.Context, offset, limit int64) ([]domain.Worker, error) {
	ts, err := repo.dao.ListWorker(ctx, offset, limit)
	return slice.Map(ts, func(idx int, src dao.Worker) domain.Worker {
		return repo.toDomain(src)
	}), err
}

func (repo *workerRepository) Total(ctx context.Context) (int64, error) {
	return repo.dao.Count(ctx)
}

func (repo *workerRepository) toEntity(req domain.Worker) dao.Worker {
	return dao.Worker{
		Name:   req.Name,
		Key:    req.Key,
		Topic:  req.Topic,
		Desc:   req.Desc,
		Status: domain.Status.ToUint8(req.Status),
	}
}

func (repo *workerRepository) toDomain(req dao.Worker) domain.Worker {
	return domain.Worker{
		Id:     req.Id,
		Key:    req.Key,
		Name:   req.Name,
		Desc:   req.Desc,
		Topic:  req.Topic,
		Status: domain.Status(req.Status),
	}
}
