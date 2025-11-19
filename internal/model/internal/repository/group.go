package repository

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/model/internal/domain"
	"github.com/Duke1616/ecmdb/internal/model/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type MGRepository interface {
	// CreateModelGroup 创建模型分组
	CreateModelGroup(ctx context.Context, req domain.ModelGroup) (int64, error)

	// BatchCreate 批量创建模型分组
	BatchCreate(ctx context.Context, req []domain.ModelGroup) ([]domain.ModelGroup, error)

	// GetByNames 根据名称查询模型组
	GetByNames(ctx context.Context, names []string) ([]domain.ModelGroup, error)

	// GetByName 根据名称查询模型组
	GetByName(ctx context.Context, name string) (domain.ModelGroup, error)

	// List 获取模型组列表
	List(ctx context.Context, offset, limit int64) ([]domain.ModelGroup, error)

	// Total 获取模型组数量
	Total(ctx context.Context) (int64, error)

	// DeleteModelGroup 根据 ID 删除模型组
	DeleteModelGroup(ctx context.Context, id int64) (int64, error)
}

type groupRepository struct {
	dao dao.ModelGroupDAO
}

func NewMGRepository(dao dao.ModelGroupDAO) MGRepository {
	return &groupRepository{
		dao: dao,
	}
}

func (repo *groupRepository) GetByName(ctx context.Context, name string) (domain.ModelGroup, error) {
	mg, err := repo.dao.GetByName(ctx, name)
	return repo.toDomain(mg), err
}

func (repo *groupRepository) BatchCreate(ctx context.Context, mgs []domain.ModelGroup) ([]domain.ModelGroup, error) {
	mgsResp, err := repo.dao.BatchCreate(ctx, slice.Map(mgs, func(idx int, src domain.ModelGroup) dao.ModelGroup {
		return dao.ModelGroup{
			Name: src.Name,
		}
	}))

	return slice.Map(mgsResp, func(idx int, src dao.ModelGroup) domain.ModelGroup {
		return repo.toDomain(src)
	}), err
}

func (repo *groupRepository) GetByNames(ctx context.Context, names []string) ([]domain.ModelGroup, error) {
	mgs, err := repo.dao.GetByNames(ctx, names)
	return slice.Map(mgs, func(idx int, src dao.ModelGroup) domain.ModelGroup {
		return repo.toDomain(src)
	}), err
}

func (repo *groupRepository) List(ctx context.Context, offset, limit int64) ([]domain.ModelGroup, error) {
	mgs, err := repo.dao.List(ctx, offset, limit)
	return slice.Map(mgs, func(idx int, src dao.ModelGroup) domain.ModelGroup {
		return repo.toDomain(src)
	}), err
}

func (repo *groupRepository) Total(ctx context.Context) (int64, error) {
	return repo.dao.Count(ctx)
}

func (repo *groupRepository) CreateModelGroup(ctx context.Context, req domain.ModelGroup) (int64, error) {
	return repo.dao.CreateModelGroup(ctx, dao.ModelGroup{
		Name: req.Name,
	})
}

func (repo *groupRepository) DeleteModelGroup(ctx context.Context, id int64) (int64, error) {
	return repo.dao.Delete(ctx, id)
}

func (repo *groupRepository) toDomain(modelDao dao.ModelGroup) domain.ModelGroup {
	return domain.ModelGroup{
		ID:   modelDao.Id,
		Name: modelDao.Name,
	}
}
