package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/domain"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
)

type AttributeRepository interface {
	CreateAttribute(ctx context.Context, req domain.Attribute) (int64, error)
	SearchAttributeFieldsByModelUid(ctx context.Context, modelUid string) ([]string, error)

	ListAttributes(ctx context.Context, modelUID string) ([]domain.Attribute, error)
	Total(ctx context.Context, modelUID string) (int64, error)

	DeleteAttribute(ctx context.Context, id int64) (int64, error)

	CustomAttributeFieldColumns(ctx *gin.Context, modelUid string, customField []string) (int64, error)
	CustomAttributeFieldColumnsReverse(ctx *gin.Context, modelUid string, customField []string) (int64, error)
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

func (a *attributeRepository) CustomAttributeFieldColumns(ctx *gin.Context, modelUid string, customField []string) (int64, error) {
	return a.dao.UpdateFieldIndex(ctx, modelUid, customField)
}

func (a *attributeRepository) CustomAttributeFieldColumnsReverse(ctx *gin.Context, modelUid string, customField []string) (int64, error) {
	return a.dao.UpdateFieldIndexReverse(ctx, modelUid, customField)
}

func (a *attributeRepository) DeleteAttribute(ctx context.Context, id int64) (int64, error) {
	return a.dao.DeleteAttribute(ctx, id)
}

func (a *attributeRepository) toEntity(req domain.Attribute) dao.Attribute {
	return dao.Attribute{
		ModelUID:  req.ModelUid,
		FieldUid:  req.FieldUid,
		FieldName: req.FieldName,
		FieldType: req.FieldType,
		Required:  req.Required,
	}
}

func (a *attributeRepository) toDomain(attr dao.Attribute) domain.Attribute {
	return domain.Attribute{
		ID:        attr.Id,
		FieldUid:  attr.FieldUid,
		FieldName: attr.FieldName,
		FieldType: attr.FieldType,
		ModelUid:  attr.ModelUID,
		Required:  attr.Required,
		Display:   attr.Display,
		Index:     attr.Index,
	}
}
