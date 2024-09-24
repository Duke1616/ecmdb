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
	UpdateUser(ctx context.Context, req domain.User) (int64, error)
	Total(ctx context.Context) (int64, error)
	AddRoleBind(ctx context.Context, id int64, roleCodes []string) (int64, error)
	UpdatePassword(ctx context.Context, id int64, password string) error
	FindByUsernameRegex(ctx context.Context, offset, limit int64, username string) ([]domain.User, error)
	TotalByUsernameRegex(ctx context.Context, username string) (int64, error)
	FindByDepartmentId(ctx context.Context, offset, limit int64, departmentId int64) ([]domain.User, error)
	TotalByDepartmentId(ctx context.Context, departmentId int64) (int64, error)
	FindByUsernames(ctx context.Context, uns []string) ([]domain.User, error)
	PipelineDepartmentId(ctx context.Context) ([]domain.UserCombination, error)
}

type userRepo struct {
	dao dao.UserDAO
}

func (repo *userRepo) FindByUsernames(ctx context.Context, uns []string) ([]domain.User, error) {
	us, err := repo.dao.FindByUsernames(ctx, uns)
	return slice.Map(us, func(idx int, src dao.User) domain.User {
		return repo.toDomain(src)
	}), err
}

func (repo *userRepo) PipelineDepartmentId(ctx context.Context) ([]domain.UserCombination, error) {
	pipeline, err := repo.dao.PipelineDepartmentId(ctx)
	return slice.Map(pipeline, func(idx int, src dao.UserPipeline) domain.UserCombination {
		return domain.UserCombination{
			DepartMentId: src.DepartMentId,
			Total:        src.Total,
			Users: slice.Map(src.Users, func(idx int, src dao.User) domain.User {
				return repo.toDomain(src)
			}),
		}
	}), err
}

func (repo *userRepo) UpdateUser(ctx context.Context, req domain.User) (int64, error) {
	return repo.dao.UpdateUser(ctx, repo.toEntity(req))
}

func (repo *userRepo) FindByDepartmentId(ctx context.Context, offset, limit int64, departmentId int64) ([]domain.User, error) {
	us, err := repo.dao.FindByDepartmentId(ctx, offset, limit, departmentId)
	return slice.Map(us, func(idx int, src dao.User) domain.User {
		return repo.toDomain(src)
	}), err
}

func (repo *userRepo) TotalByDepartmentId(ctx context.Context, departmentId int64) (int64, error) {
	return repo.dao.CountByDepartmentId(ctx, departmentId)
}

func (repo *userRepo) FindByUsernameRegex(ctx context.Context, offset, limit int64, username string) ([]domain.User, error) {
	us, err := repo.dao.FindByUsernameRegex(ctx, offset, limit, username)
	return slice.Map(us, func(idx int, src dao.User) domain.User {
		return repo.toDomain(src)
	}), err
}

func (repo *userRepo) TotalByUsernameRegex(ctx context.Context, username string) (int64, error) {
	return repo.dao.CountByUsernameRegex(ctx, username)
}

func (repo *userRepo) UpdatePassword(ctx context.Context, id int64, password string) error {
	return repo.dao.UpdatePassword(ctx, id, password)
}

func (repo *userRepo) FindById(ctx context.Context, id int64) (domain.User, error) {
	user, err := repo.dao.FindById(ctx, id)
	return repo.toDomain(user), err
}

func (repo *userRepo) AddRoleBind(ctx context.Context, id int64, roleCodes []string) (int64, error) {
	return repo.dao.AddOrUpdateRoleBind(ctx, id, roleCodes)
}

func (repo *userRepo) ListUser(ctx context.Context, offset, limit int64) ([]domain.User, error) {
	us, err := repo.dao.ListUser(ctx, offset, limit)
	return slice.Map(us, func(idx int, src dao.User) domain.User {
		return repo.toDomain(src)
	}), err
}

func (repo *userRepo) Total(ctx context.Context) (int64, error) {
	return repo.dao.Count(ctx)
}

func NewResourceRepository(dao dao.UserDAO) UserRepository {
	return &userRepo{
		dao: dao,
	}
}

func (repo *userRepo) CreatUser(ctx context.Context, user domain.User) (int64, error) {
	return repo.dao.CreatUser(ctx, repo.toEntity(user))
}

func (repo *userRepo) FindByUsername(ctx context.Context, username string) (domain.User, error) {
	user, err := repo.dao.FindByUsername(ctx, username)
	return repo.toDomain(user), err
}

func (repo *userRepo) toDomain(user dao.User) domain.User {
	return domain.User{
		Id:           user.Id,
		DepartmentId: user.DepartmentId,
		Username:     user.Username,
		RoleCodes:    user.RoleCodes,
		Password:     user.Password,
		DisplayName:  user.DisplayName,
		Email:        user.Email,
		Title:        user.Title,
		FeishuInfo: domain.FeishuInfo{
			UserId: user.FeishuInfo.UserId,
		},
		Status:     domain.Status(user.Status),
		CreateType: domain.CreateType(user.CreateType),
	}
}

func (repo *userRepo) toEntity(user domain.User) dao.User {
	return dao.User{
		Id:           user.Id,
		DepartmentId: user.DepartmentId,
		Username:     user.Username,
		Email:        user.Email,
		Title:        user.Title,
		DisplayName:  user.DisplayName,
		FeishuInfo: dao.FeishuInfo{
			UserId: user.FeishuInfo.UserId,
		},
		Status:     user.Status.ToUint8(),
		CreateType: user.CreateType.ToUint8(),
	}
}
