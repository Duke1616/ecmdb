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
	UpdatePassword(ctx context.Context, id int64, password string) error
}

type repo struct {
	dao dao.UserDAO
}

func (repo *repo) UpdatePassword(ctx context.Context, id int64, password string) error {
	return repo.dao.UpdatePassword(ctx, id, password)
}

func (repo *repo) FindById(ctx context.Context, id int64) (domain.User, error) {
	user, err := repo.dao.FindById(ctx, id)
	return repo.toDomain(user), err
}

func (repo *repo) AddRoleBind(ctx context.Context, id int64, roleCodes []string) (int64, error) {
	return repo.dao.AddOrUpdateRoleBind(ctx, id, roleCodes)
}

func (repo *repo) ListUser(ctx context.Context, offset, limit int64) ([]domain.User, error) {
	us, err := repo.dao.ListUser(ctx, offset, limit)
	return slice.Map(us, func(idx int, src dao.User) domain.User {
		return repo.toDomain(src)
	}), err
}

func (repo *repo) Total(ctx context.Context) (int64, error) {
	return repo.dao.Count(ctx)
}

func NewResourceRepository(dao dao.UserDAO) UserRepository {
	return &repo{
		dao: dao,
	}
}

func (repo *repo) CreatUser(ctx context.Context, user domain.User) (int64, error) {
	return repo.dao.CreatUser(ctx, repo.toEntity(user))
}

func (repo *repo) FindByUsername(ctx context.Context, username string) (domain.User, error) {
	user, err := repo.dao.FindByUsername(ctx, username)
	return repo.toDomain(user), err
}

func (repo *repo) toDomain(user dao.User) domain.User {
	return domain.User{
		Id:          user.Id,
		Username:    user.Username,
		RoleCodes:   user.RoleCodes,
		Password:    user.Password,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		Title:       user.Title,
		Status:      domain.Status(user.Status),
		CreateType:  domain.CreateType(user.CreateType),
	}
}

func (repo *repo) toEntity(user domain.User) dao.User {
	return dao.User{
		Username:    user.Username,
		Email:       user.Email,
		Title:       user.Title,
		DisplayName: user.DisplayName,
		Status:      user.Status.ToUint8(),
		CreateType:  user.CreateType.ToUint8(),
	}
}
