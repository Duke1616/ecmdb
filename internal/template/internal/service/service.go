package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/template/internal/domain"
	"github.com/Duke1616/ecmdb/internal/template/internal/repository"
	"github.com/Duke1616/ecmdb/pkg/hash"
	"github.com/xen0n/go-workwx"
)

type Service interface {
	FindOrCreateByWechat(ctx context.Context, req domain.WechatInfo) (domain.Template, error)
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

	t, err := s.repo.FindByHash(ctx, hash.Hash(OAInfo.TemplateContent))
	if !errors.Is(err, repository.ErrUserNotFound) {
		return t, err
	}

	t = domain.Template{
		CreateType:       domain.WechatCreate,
		Name:             req.TemplateName,
		WechatOAControls: OAInfo.TemplateContent,
		UniqueHash:       hash.Hash(OAInfo.TemplateContent),
	}

	t.Id, err = s.repo.CreateTemplate(ctx, t)
	if err != nil {
		return domain.Template{}, err
	}

	return t, nil
}
