package repostory

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/user/internal/domain"
	"github.com/Duke1616/ecmdb/internal/user/internal/repostory/dao"
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

func (repo *userRepository) CreatUser(ctx context.Context, user domain.User) (int64, error) {
	return repo.dao.CreatUser(ctx, dao.User{
		Username:   user.Username,
		Email:      user.Email,
		Title:      user.Title,
		SourceType: user.SourceType,
		CreateType: user.CreateType,
	})
}

func (repo *userRepository) FindByUsername(ctx context.Context, username string) (domain.User, error) {
	user, err := repo.dao.FindByUsername(ctx, username)
	if err != nil {
		return domain.User{}, err
	}

	return domain.User{
		ID:         user.ID,
		Username:   user.Username,
		Password:   user.Password,
		Email:      user.Email,
		Title:      user.Title,
		SourceType: user.SourceType,
		CreateType: user.CreateType,
	}, nil
}
