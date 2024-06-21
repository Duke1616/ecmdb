package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/worker/internal/domain"
	"github.com/Duke1616/ecmdb/internal/worker/internal/repository"
	"github.com/ecodeclub/mq-api"
)

type Service interface {
	CreateWorker(ctx context.Context, req domain.Worker) error
	FindOrCreateByName(ctx context.Context, req domain.Worker) (domain.Worker, error)
	ListWorker(ctx context.Context, offset, limit int64) ([]domain.Worker, error)
}

type service struct {
	repo repository.WorkerRepository
	mq   mq.MQ
}

func NewService(mq mq.MQ, repo repository.WorkerRepository) Service {
	return &service{
		mq:   mq,
		repo: repo,
	}
}

func (s *service) CreateWorker(ctx context.Context, req domain.Worker) error {
	//TODO implement me
	panic("implement me")
}

func (s *service) FindOrCreateByName(ctx context.Context, req domain.Worker) (domain.Worker, error) {
	worker, err := s.repo.FindByName(ctx, req.Name)
	if !errors.Is(err, repository.ErrUserNotFound) {
		return worker, err
	}

	_, err = s.repo.CreateWorker(ctx, req)
	if err != nil {
		return domain.Worker{}, fmt.Errorf("创建节点失败: %x", err)
	}

	if err = s.mq.CreateTopic(ctx, req.Topic, 1); err != nil {
		return domain.Worker{}, fmt.Errorf("创建Topic失败: %x", err)
	}

	return domain.Worker{}, nil
}

func (s *service) ListWorker(ctx context.Context, offset, limit int64) ([]domain.Worker, error) {
	//TODO implement me
	panic("implement me")
}
