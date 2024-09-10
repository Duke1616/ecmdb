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
	SearchAttributeFieldsBySecure(ctx context.Context, modelUid []string) (map[string][]string, error)
	ListAttributes(ctx context.Context, modelUID string) ([]domain.Attribute, error)
	Total(ctx context.Context, modelUID string) (int64, error)

	DeleteAttribute(ctx context.Context, id int64) (int64, error)

	CustomAttributeFieldColumns(ctx *gin.Context, modelUid string, customField []string) (int64, error)
	CustomAttributeFieldColumnsReverse(ctx *gin.Context, modelUid string, customField []string) (int64, error)

	ListAttributePipeline(ctx context.Context, modelUid string) ([]domain.AttributePipeline, error)
	UpdateAttribute(ctx context.Context, req domain.Attribute) (int64, error)
}

type attributeRepository struct {
	dao dao.AttributeDAO
}

func NewAttributeRepository(dao dao.AttributeDAO) AttributeRepository {
	return &attributeRepository{
		dao: dao,
	}
}

func (repo *attributeRepository) CreateAttribute(ctx context.Context, req domain.Attribute) (int64, error) {
	return repo.dao.CreateAttribute(ctx, repo.toEntity(req))
}

func (repo *attributeRepository) UpdateAttribute(ctx context.Context, req domain.Attribute) (int64, error) {
	return repo.dao.UpdateAttribute(ctx, repo.toEntity(req))
}

// SearchAttributeFieldsByModelUid 查询对应模型的字段信息
func (repo *attributeRepository) SearchAttributeFieldsByModelUid(ctx context.Context, modelUid string) ([]string, error) {
	attrs, err := repo.dao.SearchAttributeByModelUID(ctx, modelUid)

	return slice.Map(attrs, func(idx int, src dao.Attribute) string {
		return src.FieldUid
	}), err
}

func (repo *attributeRepository) ListAttributes(ctx context.Context, modelUID string) ([]domain.Attribute, error) {
	attrs, err := repo.dao.ListAttribute(ctx, modelUID)

	return slice.Map(attrs, func(idx int, src dao.Attribute) domain.Attribute {
		return repo.toDomain(src)
	}), err
}

func (repo *attributeRepository) Total(ctx context.Context, modelUID string) (int64, error) {
	return repo.dao.Count(ctx, modelUID)
}

func (repo *attributeRepository) CustomAttributeFieldColumns(ctx *gin.Context, modelUid string, customField []string) (int64, error) {
	return repo.dao.UpdateFieldIndex(ctx, modelUid, customField)
}

func (repo *attributeRepository) CustomAttributeFieldColumnsReverse(ctx *gin.Context, modelUid string, customField []string) (int64, error) {
	return repo.dao.UpdateFieldIndexReverse(ctx, modelUid, customField)
}

func (repo *attributeRepository) DeleteAttribute(ctx context.Context, id int64) (int64, error) {
	return repo.dao.DeleteAttribute(ctx, id)
}

func (repo *attributeRepository) ListAttributePipeline(ctx context.Context, modelUid string) ([]domain.AttributePipeline, error) {
	rrs, err := repo.dao.ListAttributePipeline(ctx, modelUid)
	return slice.Map(rrs, func(idx int, src dao.AttributePipeline) domain.AttributePipeline {
		return repo.toAttributeGroupsDomain(src)
	}), err
}

func (repo *attributeRepository) SearchAttributeFieldsBySecure(ctx context.Context, modelUids []string) (map[string][]string, error) {
	attrs, err := repo.dao.SearchAttributeFieldsBySecure(ctx, modelUids)
	return slice.ToMapV(attrs, func(element dao.Attribute) (string, []string) {
		return element.ModelUID, slice.FilterMap(attrs, func(idx int, src dao.Attribute) (string, bool) {
			if src.ModelUID == element.ModelUID {
				return src.FieldUid, true
			}
			return "", false
		})
	}), err
}

func (repo *attributeRepository) toEntity(req domain.Attribute) dao.Attribute {
	return dao.Attribute{
		Id:        req.ID,
		ModelUID:  req.ModelUid,
		FieldUid:  req.FieldUid,
		FieldName: req.FieldName,
		FieldType: req.FieldType,
		GroupId:   req.GroupId,
		Required:  req.Required,
		Secure:    req.Secure,
		Index:     req.Index,
		Option:    req.Option,
		Display:   req.Display,
	}
}

func (repo *attributeRepository) toDomain(attr dao.Attribute) domain.Attribute {
	return domain.Attribute{
		ID:        attr.Id,
		FieldUid:  attr.FieldUid,
		FieldName: attr.FieldName,
		FieldType: attr.FieldType,
		ModelUid:  attr.ModelUID,
		Secure:    attr.Secure,
		Required:  attr.Required,
		Option:    attr.Option,
		Display:   attr.Display,
		Index:     attr.Index,
	}
}

func (repo *attributeRepository) toAttributeGroupsDomain(ags dao.AttributePipeline) domain.AttributePipeline {
	return domain.AttributePipeline{
		GroupId: ags.GroupId,
		Total:   ags.Total,
		Attributes: slice.Map(ags.Attributes, func(idx int, src dao.Attribute) domain.Attribute {
			return repo.toDomain(src)
		}),
	}
}
