package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/domain"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/repository"
)

//go:generate mockgen -source=./service.go -destination=../../mocks/attribute.mock.go -package=attributemocks Service
type Service interface {
	CreateAttribute(ctx context.Context, req domain.Attribute) (int64, error)
	SearchAttributeFiled(ctx context.Context, modelUid string) (map[string]int, error)

	ListAttribute(ctx context.Context, modelUID string) ([]domain.Attribute, error)
}

type service struct {
	repo repository.AttributeRepository
}

func NewService(repo repository.AttributeRepository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) CreateAttribute(ctx context.Context, req domain.Attribute) (int64, error) {
	return s.repo.CreateAttribute(ctx, req)
}

func (s *service) SearchAttributeFiled(ctx context.Context, modelUid string) (map[string]int, error) {
	return s.repo.SearchAttributeByModelUID(ctx, modelUid)
}

func (s *service) ListAttribute(ctx context.Context, modelUID string) ([]domain.Attribute, error) {
	return s.repo.ListAttribute(ctx, modelUID)
}
