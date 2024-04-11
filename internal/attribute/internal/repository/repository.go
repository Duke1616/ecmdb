package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/domain"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/repository/dao"
)

type AttributeRepository interface {
	CreateAttribute(ctx context.Context, req domain.Attribute) (int64, error)
	SearchAttributeByModelIdentifies(ctx context.Context, identifies string) ([]domain.Attribute, error)
}

type attributeRepository struct {
	dao dao.AttributeDAO
}

func NewAttributeRepository(dao dao.AttributeDAO) AttributeRepository {
	return &attributeRepository{
		dao: dao,
	}
}

func (a *attributeRepository) CreateAttribute(ctx context.Context, req domain.Attribute) (int64, error) {
	return a.dao.CreateAttribute(ctx, dao.Attribute{
		ModelIdentifies: req.ModelIdentifies,
		Name:            req.Name,
		Identifies:      req.Identifies,
		FieldType:       req.FieldType,
		Required:        req.Required,
	})
}

func (a *attributeRepository) SearchAttributeByModelIdentifies(ctx context.Context, identifies string) ([]domain.Attribute, error) {
	attributeList, err := a.dao.SearchAttributeByModelIdentifies(ctx, identifies)
	if err != nil {
		return nil, err
	}

	domainAttribute := make([]domain.Attribute, 0, len(attributeList))
	for _, ca := range attributeList {
		domainAttribute = append(domainAttribute, a.toDomain(ca))
	}
	return domainAttribute, nil
}

func (a *attributeRepository) toDomain(modelDao *dao.Attribute) domain.Attribute {
	return domain.Attribute{
		ID:              modelDao.Id,
		Name:            modelDao.Name,
		Identifies:      modelDao.Identifies,
		FieldType:       modelDao.FieldType,
		ModelIdentifies: modelDao.ModelIdentifies,
		Required:        modelDao.Required,
	}
}
