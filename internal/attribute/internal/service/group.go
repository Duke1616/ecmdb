package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/domain"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/repository"
)

type GroupService interface {
	CreateAttributeGroup(ctx context.Context, req domain.AttributeGroup) (int64, error)
	ListAttributeGroup(ctx context.Context, modelUid string) ([]domain.AttributeGroup, error)
}

type groupService struct {
	repo repository.AttributeGroupRepository
}

func (g *groupService) ListAttributeGroup(ctx context.Context, modelUid string) ([]domain.AttributeGroup, error) {
	return g.repo.ListAttributeGroup(ctx, modelUid)
}

func (g *groupService) CreateAttributeGroup(ctx context.Context, req domain.AttributeGroup) (int64, error) {
	return g.repo.CreateAttributeGroup(ctx, req)
}

func NewGroupService(repo repository.AttributeGroupRepository) GroupService {
	return &groupService{
		repo: repo,
	}
}
