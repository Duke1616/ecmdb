package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/template/internal/domain"
	"github.com/Duke1616/ecmdb/internal/template/internal/repository"
	"github.com/Duke1616/ecmdb/pkg/hash"
	"github.com/xen0n/go-workwx"
	"golang.org/x/sync/errgroup"
)

//go:generate mockgen -source=./service.go -package=templatemocks -destination=../../mocks/template.mock.go -typed Service
type Service interface {
	// FindOrCreateByWechat 创建或查询来自企业微信的 OA 模版，会把企业微信的 OA 模版同步到本系统存储
	FindOrCreateByWechat(ctx context.Context, req domain.WechatInfo) (domain.Template, error)
	// CreateTemplate 创建模版，并返回模版 ID
	CreateTemplate(ctx context.Context, req domain.Template) (int64, error)
	// DetailTemplate 获取对应 ID 的模版详情配置
	DetailTemplate(ctx context.Context, id int64) (domain.Template, error)
	// FindByTemplateIds 根据一系列 ID 获取对应的模版领域模型列表
	FindByTemplateIds(ctx context.Context, ids []int64) ([]domain.Template, error)
	// DetailTemplateByExternalTemplateId 通过外部关联系统 ID (如飞书、企业微信) 获取对应系统内对应模版详情
	DetailTemplateByExternalTemplateId(ctx context.Context, externalId string) (domain.Template, error)
	// ListTemplate 分页获取所有模版列表并返回包含所有模版的合计数目
	ListTemplate(ctx context.Context, offset, limit int64) ([]domain.Template, int64, error)
	// DeleteTemplate 删除指定的模版对象，以 ID 标识返回被影响记录数目
	DeleteTemplate(ctx context.Context, id int64) (int64, error)
	// UpdateTemplate 更新替换既有的模版属性
	UpdateTemplate(ctx context.Context, t domain.Template) (int64, error)
	// Pipeline 负责以组别结构的形式聚合返回不同组合的模版列表以供工作流或者前端选项调起展示
	Pipeline(ctx context.Context) ([]domain.TemplateCombination, error)
	// GetByWorkflowId 获取同一业务线上某个具体工作流程上归属管理下的不同模版清单
	GetByWorkflowId(ctx context.Context, workflowId int64) ([]domain.Template, error)
	// FindByKeyword 以特定输入词执行模版的查询过滤并且以分页机制反馈清单与它的计数汇总条数
	FindByKeyword(ctx context.Context, keyword string, offset, limit int64) ([]domain.Template, int64, error)
	// ToggleFavorite 为指定用户触发针对单模板配置实体验证的添加与解除关联状态动作并即时返回新近确立后的最新属性值
	ToggleFavorite(ctx context.Context, userId int64, templateId int64) (bool, error)
	// ListFavoriteTemplates 呈现指定关联用户曾经标定过并置于专属藏品当中的系列模版整体汇总状态目录
	ListFavoriteTemplates(ctx context.Context, userId int64) ([]domain.Template, error)
}

type service struct {
	repo    repository.TemplateRepository
	workApp *workwx.WorkwxApp
}

func (s *service) ToggleFavorite(ctx context.Context, userId int64, templateId int64) (bool, error) {
	return s.repo.ToggleFavorite(ctx, userId, templateId)
}

func (s *service) ListFavoriteTemplates(ctx context.Context, userId int64) ([]domain.Template, error) {
	return s.repo.ListFavoriteTemplates(ctx, userId)
}

func (s *service) GetByWorkflowId(ctx context.Context, workflowId int64) ([]domain.Template, error) {
	return s.repo.GetByWorkflowId(ctx, workflowId)
}

func (s *service) DetailTemplateByExternalTemplateId(ctx context.Context, externalId string) (domain.Template, error) {
	return s.repo.DetailTemplateByExternalTemplateId(ctx, externalId)
}

func (s *service) FindByTemplateIds(ctx context.Context, ids []int64) ([]domain.Template, error) {
	return s.repo.FindByTemplateIds(ctx, ids)
}

func NewService(repo repository.TemplateRepository, workApp *workwx.WorkwxApp) Service {
	return &service{
		repo:    repo,
		workApp: workApp,
	}
}

func (s *service) FindOrCreateByWechat(ctx context.Context, req domain.WechatInfo) (domain.Template, error) {
	OAInfo, err := s.workApp.GetOATemplateDetail(req.TemplateId)
	if err != nil {
		return domain.Template{}, fmt.Errorf("获取模版详情失败: %w", err)
	}

	t, err := s.repo.FindByExternalTemplateId(ctx, req.TemplateId)
	if !errors.Is(err, repository.ErrUserNotFound) {
		if hash.Hash(OAInfo.TemplateContent) != hash.Hash(t.WechatOAControls) {
			// TODO 重新同步 Controls 数据

		}

		return t, err
	}

	t = domain.Template{
		CreateType:         domain.WechatCreate,
		Name:               req.TemplateName,
		ExternalTemplateId: req.TemplateId,
		WechatOAControls:   OAInfo.TemplateContent,
		UniqueHash:         hash.Hash(OAInfo.TemplateContent),
	}

	t.Id, err = s.repo.CreateTemplate(ctx, t)
	if err != nil {
		return domain.Template{}, err
	}

	return t, nil
}

func (s *service) CreateTemplate(ctx context.Context, req domain.Template) (int64, error) {
	return s.repo.CreateTemplate(ctx, req)
}

func (s *service) UpdateTemplate(ctx context.Context, t domain.Template) (int64, error) {
	return s.repo.UpdateTemplate(ctx, t)
}

func (s *service) DetailTemplate(ctx context.Context, id int64) (domain.Template, error) {
	return s.repo.DetailTemplate(ctx, id)
}

func (s *service) ListTemplate(ctx context.Context, offset, limit int64) ([]domain.Template, int64, error) {
	var (
		eg    errgroup.Group
		ts    []domain.Template
		total int64
	)
	eg.Go(func() error {
		var err error
		ts, err = s.repo.ListTemplate(ctx, offset, limit)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.Total(ctx)
		return err
	})
	if err := eg.Wait(); err != nil {
		return ts, total, err
	}
	return ts, total, nil
}

func (s *service) DeleteTemplate(ctx context.Context, id int64) (int64, error) {
	return s.repo.DeleteTemplate(ctx, id)
}

func (s *service) Pipeline(ctx context.Context) ([]domain.TemplateCombination, error) {
	return s.repo.Pipeline(ctx)
}

func (s *service) FindByKeyword(ctx context.Context, keyword string, offset, limit int64) ([]domain.Template, int64, error) {
	var (
		eg    errgroup.Group
		ts    []domain.Template
		total int64
	)
	eg.Go(func() error {
		var err error
		ts, err = s.repo.FindByKeyword(ctx, keyword, offset, limit)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.CountByKeyword(ctx, keyword)
		return err
	})
	if err := eg.Wait(); err != nil {
		return ts, total, err
	}
	return ts, total, nil
}
