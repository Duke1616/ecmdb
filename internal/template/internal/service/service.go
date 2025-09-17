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

	// CreateTemplate 创建模版
	CreateTemplate(ctx context.Context, req domain.Template) (int64, error)

	// DetailTemplate 查询模版详情
	DetailTemplate(ctx context.Context, id int64) (domain.Template, error)

	// FindByTemplateIds 根据 IDS 获取模版列表
	FindByTemplateIds(ctx context.Context, ids []int64) ([]domain.Template, error)

	// DetailTemplateByExternalTemplateId 查询外部模版，比如集成飞书OA之类的，统一一个ID
	DetailTemplateByExternalTemplateId(ctx context.Context, externalId string) (domain.Template, error)

	// ListTemplate 获取模版列表
	ListTemplate(ctx context.Context, offset, limit int64) ([]domain.Template, int64, error)

	// DeleteTemplate 删除模版
	DeleteTemplate(ctx context.Context, id int64) (int64, error)

	// UpdateTemplate 修改模版
	UpdateTemplate(ctx context.Context, t domain.Template) (int64, error)

	// Pipeline 返回每个组下的模版列表
	Pipeline(ctx context.Context) ([]domain.TemplateCombination, error)

	// GetByWorkflowId 根据流程ID，获取绑定的模版列表
	GetByWorkflowId(ctx context.Context, workflowId int64) ([]domain.Template, error)
	// 创建流程，生成传递数据的规则
}

type service struct {
	repo    repository.TemplateRepository
	workApp *workwx.WorkwxApp
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
