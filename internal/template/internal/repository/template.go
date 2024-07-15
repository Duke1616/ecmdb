package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/template/internal/domain"
	"github.com/Duke1616/ecmdb/internal/template/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type TemplateRepository interface {
	CreateTemplate(ctx context.Context, req domain.Template) (int64, error)
	FindByHash(ctx context.Context, hash string) (domain.Template, error)
	FindByExternalTemplateId(ctx context.Context, hash string) (domain.Template, error)
	DetailTemplate(ctx context.Context, id int64) (domain.Template, error)
	DeleteTemplate(ctx context.Context, id int64) (int64, error)
	UpdateTemplate(ctx context.Context, req domain.Template) (int64, error)
	ListTemplate(ctx context.Context, offset, limit int64) ([]domain.Template, error)
	Total(ctx context.Context) (int64, error)
	Pipeline(ctx context.Context) ([]domain.TemplateCombination, error)
}

func NewTemplateRepository(dao dao.TemplateDAO) TemplateRepository {
	return &templateRepository{
		dao: dao,
	}
}

type templateRepository struct {
	dao dao.TemplateDAO
}

func (repo *templateRepository) Pipeline(ctx context.Context) ([]domain.TemplateCombination, error) {
	pipeline, err := repo.dao.Pipeline(ctx)
	return slice.Map(pipeline, func(idx int, src dao.TemplatePipeline) domain.TemplateCombination {
		return domain.TemplateCombination{
			Id:    src.Id,
			Total: src.Total,
			Templates: slice.Map(src.Templates, func(idx int, src dao.Template) domain.Template {
				return repo.toDomain(src)
			}),
		}
	}), err
}

func (repo *templateRepository) CreateTemplate(ctx context.Context, req domain.Template) (int64, error) {
	return repo.dao.CreateTemplate(ctx, repo.toEntity(req))
}

func (repo *templateRepository) FindByHash(ctx context.Context, hash string) (domain.Template, error) {
	t, err := repo.dao.FindByHash(ctx, hash)
	return repo.toDomain(t), err
}

func (repo *templateRepository) FindByExternalTemplateId(ctx context.Context, externalTemplateId string) (domain.Template, error) {
	t, err := repo.dao.FindByExternalTemplateId(ctx, externalTemplateId)
	return repo.toDomain(t), err
}

func (repo *templateRepository) DetailTemplate(ctx context.Context, id int64) (domain.Template, error) {
	t, err := repo.dao.DetailTemplate(ctx, id)
	return repo.toDomain(t), err
}

func (repo *templateRepository) DeleteTemplate(ctx context.Context, id int64) (int64, error) {
	return repo.dao.DeleteTemplate(ctx, id)
}

func (repo *templateRepository) UpdateTemplate(ctx context.Context, req domain.Template) (int64, error) {
	return repo.dao.UpdateTemplate(ctx, repo.toEntity(req))
}

func (repo *templateRepository) ListTemplate(ctx context.Context, offset, limit int64) ([]domain.Template, error) {
	ts, err := repo.dao.ListTemplate(ctx, offset, limit)
	return slice.Map(ts, func(idx int, src dao.Template) domain.Template {
		return repo.toDomain(src)
	}), err
}

func (repo *templateRepository) Total(ctx context.Context) (int64, error) {
	return repo.dao.Count(ctx)
}

func (repo *templateRepository) toEntity(req domain.Template) dao.Template {
	return dao.Template{
		Id:                 req.Id,
		Name:               req.Name,
		WorkflowId:         req.WorkflowId,
		GroupId:            req.GroupId,
		Icon:               req.Icon,
		CreateType:         req.CreateType.ToUint8(),
		UniqueHash:         req.UniqueHash,
		WechatOAControls:   req.WechatOAControls,
		ExternalTemplateId: req.ExternalTemplateId,
		Rules:              req.Rules,
		Options:            req.Options,
		Desc:               req.Desc,
	}
}

func (repo *templateRepository) toDomain(req dao.Template) domain.Template {
	return domain.Template{
		Id:                 req.Id,
		Name:               req.Name,
		WorkflowId:         req.WorkflowId,
		GroupId:            req.GroupId,
		Icon:               req.Icon,
		CreateType:         domain.CreateType(req.CreateType),
		WechatOAControls:   req.WechatOAControls,
		UniqueHash:         req.UniqueHash,
		ExternalTemplateId: req.ExternalTemplateId,
		Rules:              req.Rules,
		Options:            req.Options,
		Desc:               req.Desc,
	}
}
