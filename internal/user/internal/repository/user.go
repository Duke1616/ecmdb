package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/user/internal/domain"
	"github.com/Duke1616/ecmdb/internal/user/internal/repository/dao"
)

type UserRepository interface {
	CreatUser(ctx context.Context, user domain.User) (int64, error)
	FindByUsername(ctx context.Context, username string) (domain.User, error)
}

type userRepository struct {
	dao dao.UserDAO
}

func NewResourceRepository(dao dao.UserDAO) UserRepository {
	return &userRepository{
		dao: dao,
	}
}

func (r *userRepository) CreatUser(ctx context.Context, user domain.User) (int64, error) {
	return r.dao.CreatUser(ctx, r.toEntity(user))
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (domain.User, error) {
	user, err := r.dao.FindByUsername(ctx, username)
	return r.toDomain(user), err
}

func (r *userRepository) toDomain(user dao.User) domain.User {
	return domain.User{
		ID:         user.ID,
		Username:   user.Username,
		Password:   user.Password,
		Email:      user.Email,
		Title:      user.Title,
		SourceType: user.SourceType,
		CreateType: user.CreateType,
	}
}

func (r *userRepository) toEntity(user domain.User) dao.User {
	return dao.User{
		Username:   user.Username,
		Email:      user.Email,
		Title:      user.Title,
		SourceType: user.SourceType,
		CreateType: user.CreateType,
	}
}
