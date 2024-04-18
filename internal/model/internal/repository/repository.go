package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/model/internal/domain"
	"github.com/Duke1616/ecmdb/internal/model/internal/repository/dao"
	"time"
)

type ModelRepository interface {
	CreateModelGroup(ctx context.Context, req domain.ModelGroup) (int64, error)
	CreateModel(ctx context.Context, req domain.Model) (int64, error)
	FindModelByUid(ctx context.Context, Identifies string) (domain.Model, error)
	ListModels(ctx context.Context, offset, limit int64) ([]domain.Model, error)
	Total(ctx context.Context) (int64, error)
}

func NewModelRepository(dao dao.ModelDAO) ModelRepository {
	return &modelRepository{
		dao: dao,
	}
}

type modelRepository struct {
	dao dao.ModelDAO
}

func (m *modelRepository) Total(ctx context.Context) (int64, error) {
	return m.dao.CountModels(ctx)
}

func (m *modelRepository) ListModels(ctx context.Context, offset, limit int64) ([]domain.Model, error) {
	modelList, err := m.dao.ListModels(ctx, offset, limit)
	if err != nil {
		return nil, err
	}

	domainModels := make([]domain.Model, 0, len(modelList))
	for _, ca := range modelList {
		domainModels = append(domainModels, m.toDomain(ca))
	}
	return domainModels, nil
}

func (m *modelRepository) CreateModel(ctx context.Context, req domain.Model) (int64, error) {
	return m.dao.CreateModel(ctx, dao.Model{
		ModelGroupId: req.GroupId,
		Name:         req.Name,
		UID:          req.UID,
		Icon:         req.Icon,
	})
}

func (m *modelRepository) CreateModelGroup(ctx context.Context, req domain.ModelGroup) (int64, error) {
	return m.dao.CreateModelGroup(ctx, dao.ModelGroup{
		Name: req.Name,
	})
}

func (m *modelRepository) FindModelByUid(ctx context.Context, uid string) (domain.Model, error) {
	model, err := m.dao.GetModelByUid(ctx, uid)
	if err != nil {
		return domain.Model{}, err
	}

	return domain.Model{
		ID:      model.Id,
		GroupId: model.ModelGroupId,
		Name:    model.Name,
		UID:     model.UID,
	}, nil
}

func (m *modelRepository) toDomain(modelDao dao.Model) domain.Model {
	return domain.Model{
		ID:      modelDao.Id,
		GroupId: modelDao.ModelGroupId,
		Name:    modelDao.Name,
		UID:     modelDao.UID,
		Icon:    modelDao.Icon,
		Ctime:   time.UnixMilli(modelDao.Ctime),
		Utime:   time.UnixMilli(modelDao.Utime),
	}
}
