package repository

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/template/internal/domain"
	"github.com/Duke1616/ecmdb/internal/template/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type TemplateRepository interface {
	// CreateTemplate 在数据层创建模板，并返回生成的 ID
	CreateTemplate(ctx context.Context, req domain.Template) (int64, error)
	// FindByHash 通过哈希值获取模板领域模型，用于判断模板是否重复
	FindByHash(ctx context.Context, hash string) (domain.Template, error)
	// FindByExternalTemplateId 通过外部模板 ID 获取模板信息
	FindByExternalTemplateId(ctx context.Context, hash string) (domain.Template, error)
	// DetailTemplate 获取对应 ID 的模板详情
	DetailTemplate(ctx context.Context, id int64) (domain.Template, error)
	// DeleteTemplate 删除指定的模板并返回删除的记录数
	DeleteTemplate(ctx context.Context, id int64) (int64, error)
	// DetailTemplateByExternalTemplateId 通过外部模板 ID 获取对应的详细信息
	DetailTemplateByExternalTemplateId(ctx context.Context, externalId string) (domain.Template, error)
	// UpdateTemplate 更新数据层中的模板相关信息
	UpdateTemplate(ctx context.Context, req domain.Template) (int64, error)
	// ListTemplate 根据分页条件获取模板列表
	ListTemplate(ctx context.Context, offset, limit int64) ([]domain.Template, error)
	// Total 统计模板总行数
	Total(ctx context.Context) (int64, error)
	// Pipeline 返回根据分组整理的分类模板数据，常用于前端分组选单展示
	Pipeline(ctx context.Context) ([]domain.TemplateCombination, error)

	// FindByTemplateIds 根据一批模板 ID 拉取对应的模板详情集合
	FindByTemplateIds(ctx context.Context, ids []int64) ([]domain.Template, error)

	// GetByWorkflowId 获取某个特定工作流关联的所有模板
	GetByWorkflowId(ctx context.Context, workflowId int64) ([]domain.Template, error)

	// FindByKeyword 按照关键字进行模板搜索操作并返回分页数据
	FindByKeyword(ctx context.Context, keyword string, offset, limit int64) ([]domain.Template, error)
	// CountByKeyword 计算含有对应关键字特征的模板总记录数
	CountByKeyword(ctx context.Context, keyword string) (int64, error)

	// ToggleFavorite 切换目标模板的收藏状态（收藏/取消收藏）, 并返回其是否在操作后处于被收藏的布尔状态
	ToggleFavorite(ctx context.Context, userId int64, templateId int64) (bool, error)
	// ListFavoriteTemplates 拉取特定用户的全套模板藏品集合列
	ListFavoriteTemplates(ctx context.Context, userId int64) ([]domain.Template, error)
}

func NewTemplateRepository(dao dao.TemplateDAO, favorite dao.FavoriteDAO) TemplateRepository {
	return &templateRepository{
		dao:      dao,
		favorite: favorite,
	}
}

type templateRepository struct {
	dao      dao.TemplateDAO
	favorite dao.FavoriteDAO
}

func (repo *templateRepository) ToggleFavorite(ctx context.Context, userId int64, templateId int64) (bool, error) {
	return repo.favorite.ToggleFavorite(ctx, userId, templateId)
}

func (repo *templateRepository) ListFavoriteTemplates(ctx context.Context, userId int64) ([]domain.Template, error) {
	ids, err := repo.favorite.ListTemplateIdsByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}

	return repo.FindByTemplateIds(ctx, ids)
}

func (repo *templateRepository) FindByTemplateIds(ctx context.Context, ids []int64) ([]domain.Template, error) {
	us, err := repo.dao.FindByTemplateIds(ctx, ids)
	return slice.Map(us, func(idx int, src dao.Template) domain.Template {
		return repo.toDomain(src)
	}), err
}

func (repo *templateRepository) GetByWorkflowId(ctx context.Context, workflowId int64) ([]domain.Template, error) {
	ts, err := repo.dao.GetByWorkflowId(ctx, workflowId)
	return slice.Map(ts, func(idx int, src dao.Template) domain.Template {
		return repo.toDomain(src)
	}), err
}

func (repo *templateRepository) DetailTemplateByExternalTemplateId(ctx context.Context, externalId string) (
	domain.Template, error) {
	t, err := repo.dao.DetailTemplateByExternalTemplateId(ctx, externalId)
	return repo.toDomain(t), err
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

func (repo *templateRepository) FindByKeyword(ctx context.Context, keyword string, offset, limit int64) ([]domain.Template, error) {
	ts, err := repo.dao.FindByKeyword(ctx, keyword, offset, limit)
	return slice.Map(ts, func(idx int, src dao.Template) domain.Template {
		return repo.toDomain(src)
	}), err
}

func (repo *templateRepository) CountByKeyword(ctx context.Context, keyword string) (int64, error) {
	return repo.dao.CountByKeyword(ctx, keyword)
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
