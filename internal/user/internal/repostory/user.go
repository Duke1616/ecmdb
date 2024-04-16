package repostory

import "github.com/Duke1616/ecmdb/internal/user/internal/repostory/dao"

type UserRepository interface {
}

type userRepository struct {
	dao dao.UserDAO
}

func NewResourceRepository(dao dao.UserDAO) UserRepository {
	return &userRepository{
		dao: dao,
	}
}
