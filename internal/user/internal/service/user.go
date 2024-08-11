package service

import (
	"context"
	"errors"
	"github.com/Duke1616/ecmdb/internal/user/internal/domain"
	"github.com/Duke1616/ecmdb/internal/user/internal/repository"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	FindOrCreateByUsername(ctx context.Context, user domain.User) (domain.User, error)
	ListUser(ctx context.Context, offset, limit int64) ([]domain.User, int64, error)
	AddRoleBind(ctx context.Context, id int64, roleCodes []string) (int64, error)
}

type service struct {
	repo repository.UserRepository
}

func (s *service) AddRoleBind(ctx context.Context, id int64, roleCodes []string) (int64, error) {
	return s.repo.AddRoleBind(ctx, id, roleCodes)
}

func (s *service) ListUser(ctx context.Context, offset, limit int64) ([]domain.User, int64, error) {
	var (
		eg    errgroup.Group
		us    []domain.User
		total int64
	)
	eg.Go(func() error {
		var err error
		us, err = s.repo.ListUser(ctx, offset, limit)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.Total(ctx)
		return err
	})
	if err := eg.Wait(); err != nil {
		return us, total, err
	}
	return us, total, nil
}

func NewService(repo repository.UserRepository) Service {
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
