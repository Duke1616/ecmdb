package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/template/internal/domain"
	"github.com/Duke1616/ecmdb/internal/template/internal/repository"
)

type Service interface {
	FindOrCreateTemplate(ctx context.Context, req domain.Template) (domain.Template, error)
}

type service struct {
	repo repository.TemplateRepository
}

func NewService(repo repository.TemplateRepository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) FindOrCreateTemplate(ctx context.Context, req domain.Template) (domain.Template, error) {
	t, err := s.repo.FindByHash(ctx, req.UniqueHash)
	if !errors.Is(err, repository.ErrUserNotFound) {
		fmt.Println(t, "系统存在了")
		return t, err
	}

	return req, s.repo.CreateTemplate(ctx, req)
}
