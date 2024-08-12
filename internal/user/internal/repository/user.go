package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/user/internal/domain"
	"github.com/Duke1616/ecmdb/internal/user/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type UserRepository interface {
	CreatUser(ctx context.Context, user domain.User) (int64, error)
	FindByUsername(ctx context.Context, username string) (domain.User, error)
	FindById(ctx context.Context, id int64) (domain.User, error)
	ListUser(ctx context.Context, offset, limit int64) ([]domain.User, error)
	Total(ctx context.Context) (int64, error)
	AddRoleBind(ctx context.Context, id int64, roleCodes []string) (int64, error)
}

type userRepository struct {
	dao dao.UserDAO
}

func (repo *userRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	user, err := repo.dao.FindById(ctx, id)
	return repo.toDomain(user), err
}

func (repo *userRepository) AddRoleBind(ctx context.Context, id int64, roleCodes []string) (int64, error) {
	return repo.dao.AddOrUpdateRoleBind(ctx, id, roleCodes)
}

func (repo *userRepository) ListUser(ctx context.Context, offset, limit int64) ([]domain.User, error) {
	us, err := repo.dao.ListUser(ctx, offset, limit)
	return slice.Map(us, func(idx int, src dao.User) domain.User {
		return repo.toDomain(src)
	}), err
}

func (repo *userRepository) Total(ctx context.Context) (int64, error) {
	return repo.dao.Count(ctx)
}

func NewResourceRepository(dao dao.UserDAO) UserRepository {
	return &userRepository{
		dao: dao,
	}
}

func (repo *userRepository) CreatUser(ctx context.Context, user domain.User) (int64, error) {
	return repo.dao.CreatUser(ctx, repo.toEntity(user))
}

func (repo *userRepository) FindByUsername(ctx context.Context, username string) (domain.User, error) {
	user, err := repo.dao.FindByUsername(ctx, username)
	return repo.toDomain(user), err
}

func (repo *userRepository) toDomain(user dao.User) domain.User {
	return domain.User{
		ID:         user.ID,
		Username:   user.Username,
		RoleCodes:  user.RoleCodes,
		Password:   user.Password,
		Email:      user.Email,
		Title:      user.Title,
		SourceType: user.SourceType,
		CreateType: user.CreateType,
	}
}

func (repo *userRepository) toEntity(user domain.User) dao.User {
	return dao.User{
		Username:   user.Username,
		Email:      user.Email,
		Title:      user.Title,
		SourceType: user.SourceType,
		CreateType: user.CreateType,
	}
}
