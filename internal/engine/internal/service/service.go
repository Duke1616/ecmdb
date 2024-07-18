package service

import (
	"context"
	"github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/Duke1616/ecmdb/internal/engine/internal/domain"
	"github.com/Duke1616/ecmdb/internal/engine/internal/repository"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	ListTodo(ctx context.Context, userId, processName string, sortByAse bool, offset, limit int) (
		[]domain.Instance, int64, error)
	ListByStartUser(ctx context.Context, userId, processName string, offset, limit int) (
		[]domain.Instance, int64, error)
	TaskRecord(ctx context.Context, processInstId, offset, limit int) ([]model.Task, int64, error)
	Pass(ctx context.Context, taskId int, comment string) error
	Reject(ctx context.Context, taskId int, comment string) error
}

type service struct {
	repo repository.ProcessEngineRepository
}

func (s *service) Reject(ctx context.Context, taskId int, comment string) error {
	return engine.TaskReject(taskId, comment, "")
}

func (s *service) Pass(ctx context.Context, taskId int, comment string) error {
	return engine.TaskPass(taskId, comment, "", false)
}

func (s *service) TaskRecord(ctx context.Context, processInstId, offset, limit int) ([]model.Task, int64, error) {
	var (
		eg      errgroup.Group
		records []model.Task
		total   int64
	)
	eg.Go(func() error {
		var err error
		records, err = s.repo.ListTaskRecord(ctx, processInstId, offset, limit)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.CountTaskRecord(ctx, processInstId)
		return err
	})
	if err := eg.Wait(); err != nil {
		return records, total, err
	}
	return records, total, nil
}

func NewService(repo repository.ProcessEngineRepository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) ListTodo(ctx context.Context, userId, processName string, sortByAse bool, offset, limit int) (
	[]domain.Instance, int64, error) {
	var (
		eg    errgroup.Group
		ts    []domain.Instance
		total int64
	)
	eg.Go(func() error {
		var err error
		ts, err = s.repo.ListTodoList(userId, processName, sortByAse, offset, limit)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.CountTodo(ctx, userId, processName)
		return err
	})
	if err := eg.Wait(); err != nil {
		return ts, total, err
	}
	return ts, total, nil
}

func (s *service) ListByStartUser(ctx context.Context, userId, processName string, offset,
	limit int) ([]domain.Instance, int64, error) {

	var (
		eg    errgroup.Group
		ts    []domain.Instance
		total int64
	)
	eg.Go(func() error {
		var err error
		ts, err = s.repo.ListStartUser(ctx, userId, processName, offset, limit)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.CountStartUser(ctx, userId, processName)
		return err
	})
	if err := eg.Wait(); err != nil {
		return ts, total, err
	}
	return ts, total, nil
}
