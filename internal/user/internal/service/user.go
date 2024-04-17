package service

import (
	"context"
	"errors"
	"github.com/Duke1616/ecmdb/internal/user/internal/domain"
	"github.com/Duke1616/ecmdb/internal/user/internal/repostory"
	"go.mongodb.org/mongo-driver/mongo"
)

type Service interface {
	FindOrCreateByUsername(ctx context.Context, user domain.User) (domain.User, error)
}

type service struct {
	repo repostory.UserRepository
}

func NewService(repo repostory.UserRepository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) FindOrCreateByUsername(ctx context.Context, user domain.User) (domain.User, error) {
	u, err := s.repo.FindByUsername(ctx, user.Username)
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return u, err
	}

	id, err := s.repo.CreatUser(ctx, user)
	if err != nil {
		return domain.User{}, err
	}

	return domain.User{
		ID:         id,
		Username:   user.Username,
		Email:      user.Email,
		Title:      user.Title,
		SourceType: domain.Ldap,
		CreateType: domain.UserRegistry,
	}, err
}
