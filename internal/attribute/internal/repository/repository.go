package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/domain"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type AttributeRepository interface {
	CreateAttribute(ctx context.Context, req domain.Attribute) (int64, error)
	SearchAttributeFieldsByModelUid(ctx context.Context, modelUid string) ([]string, error)

	ListAttributes(ctx context.Context, modelUID string) ([]domain.Attribute, error)
	Total(ctx context.Context, modelUID string) (int64, error)
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
	return a.dao.CreateAttribute(ctx, a.toEntity(req))
}

// SearchAttributeFieldsByModelUid 查询对应模型的字段信息
func (a *attributeRepository) SearchAttributeFieldsByModelUid(ctx context.Context, modelUid string) ([]string, error) {
	attrs, err := a.dao.SearchAttributeByModelUID(ctx, modelUid)

	return slice.Map(attrs, func(idx int, src dao.Attribute) string {
		return src.FieldName
	}), err
}

func (a *attributeRepository) ListAttributes(ctx context.Context, modelUID string) ([]domain.Attribute, error) {
	attrs, err := a.dao.ListAttribute(ctx, modelUID)

	return slice.Map(attrs, func(idx int, src dao.Attribute) domain.Attribute {
		return a.toDomain(src)
	}), err
}

func (a *attributeRepository) Total(ctx context.Context, modelUID string) (int64, error) {
	return a.dao.Count(ctx, modelUID)
}

func (a *attributeRepository) toEntity(req domain.Attribute) dao.Attribute {
	return dao.Attribute{
		ModelUID:  req.ModelUID,
		Name:      req.Name,
		FieldName: req.FieldName,
		FieldType: req.FieldType,
		Required:  req.Required,
	}
}

func (a *attributeRepository) toDomain(attr dao.Attribute) domain.Attribute {
	return domain.Attribute{
		ID:        attr.Id,
		Name:      attr.Name,
		FieldName: attr.FieldName,
		FieldType: attr.FieldType,
		ModelUID:  attr.ModelUID,
		Required:  attr.Required,
	}
}
