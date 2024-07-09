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

type Service interface {
	FindOrCreateByWechat(ctx context.Context, req domain.WechatInfo) (domain.Template, error)
	CreateTemplate(ctx context.Context, req domain.Template) (int64, error)
	DetailTemplate(ctx context.Context, id int64) (domain.Template, error)
	ListTemplate(ctx context.Context, offset, limit int64) ([]domain.Template, int64, error)
	DeleteTemplate(ctx context.Context, id int64) (int64, error)
	UpdateTemplate(ctx context.Context, t domain.Template) (int64, error)
}

type service struct {
	repo    repository.TemplateRepository
	workApp *workwx.WorkwxApp
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
