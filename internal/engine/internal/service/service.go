package service

import (
	"context"
	"github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/Duke1616/ecmdb/internal/engine/internal/repository"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	ListTodo(ctx context.Context, userId, processName string, sortByAse bool, offset, limit int) (
		[]model.Task, int64, error)
	Pass(ctx context.Context, taskId int64)
}

type service struct {
	repo repository.ProcessEngineRepository
}

func NewService(repo repository.ProcessEngineRepository) Service {
	return &service{
		repo: repo,
	}
}

//func (s *service) UpdateOrderStatus(ctx context.Context, instanceId int, status uint8) error {
//	return s.orderSvc.UpdateStatusByInstanceId(ctx, instanceId, status)
//}

func (s *service) ListTodo(ctx context.Context, userId, processName string, sortByAse bool, offset, limit int) (
	[]model.Task, int64, error) {
	var (
		eg    errgroup.Group
		ts    []model.Task
		total int64
	)
	eg.Go(func() error {
		var err error
		ts, err = engine.GetTaskToDoList(userId, processName, sortByAse, offset, limit)
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

func (s *service) Pass(ctx context.Context, taskId int64) {
	//TODO implement me
	panic("implement me")
}
