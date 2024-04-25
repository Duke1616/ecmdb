package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/model/internal/domain"
	"github.com/Duke1616/ecmdb/internal/model/internal/repository"
)

type MGService interface {
	CreateModelGroup(ctx context.Context, req domain.ModelGroup) (int64, error)
}

type groupService struct {
	repo repository.MGRepository
}

func NewMGService(repo repository.MGRepository) MGService {
	return &groupService{
		repo: repo,
	}
}

func (s *groupService) CreateModelGroup(ctx context.Context, req domain.ModelGroup) (int64, error) {
	return s.repo.CreateModelGroup(ctx, req)
}
