package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/template/internal/domain"
	"github.com/Duke1616/ecmdb/internal/template/internal/repository/dao"
)

type TemplateRepository interface {
	CreateTemplate(ctx context.Context, req domain.Template) (int64, error)
	FindByHash(ctx context.Context, hash string) (domain.Template, error)
}

func NewTemplateRepository(dao dao.TemplateDAO) TemplateRepository {
	return &templateRepository{
		dao: dao,
	}
}

type templateRepository struct {
	dao dao.TemplateDAO
}

func (repo *templateRepository) CreateTemplate(ctx context.Context, req domain.Template) (int64, error) {
	return repo.dao.CreateTemplate(ctx, repo.toEntity(req))
}

func (repo *templateRepository) FindByHash(ctx context.Context, hash string) (domain.Template, error) {
	t, err := repo.dao.FindByHash(ctx, hash)
	return repo.toDomain(t), err
}

func (repo *templateRepository) toEntity(req domain.Template) dao.Template {
	return dao.Template{
		Name:             req.Name,
		CreateType:       req.CreateType.ToUint8(),
		UniqueHash:       req.UniqueHash,
		WechatOAControls: req.WechatOAControls,
		Rules:            req.Rules,
		Options:          req.Options,
	}
}

func (repo *templateRepository) toDomain(req dao.Template) domain.Template {
	return domain.Template{
		Id:               req.Id,
		CreateType:       domain.CreateType(req.CreateType),
		WechatOAControls: req.WechatOAControls,
		UniqueHash:       req.UniqueHash,
		Rules:            req.Rules,
		Options:          req.Options,
	}
}
