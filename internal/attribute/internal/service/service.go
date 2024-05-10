package service

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/domain"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/repository"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

//go:generate mockgen -source=./service.go -destination=../../mocks/attribute.mock.go -package=attributemocks -typed Service
type Service interface {
	CreateAttribute(ctx context.Context, req domain.Attribute) (int64, error)
	// SearchAttributeFieldsByModelUid 查询模型下的所有字段信息，内部使用
	SearchAttributeFieldsByModelUid(ctx context.Context, modelUid string) ([]string, error)
	// ListAttributes 查询模型下的所有字段详情信息，前端使用
	ListAttributes(ctx context.Context, modelUID string) ([]domain.Attribute, int64, error)

	// CustomAttributeFieldColumns 自定义展示字段、以及排序
	CustomAttributeFieldColumns(ctx *gin.Context, modelUid string, customField []string) (int64, error)
}

type service struct {
	repo repository.AttributeRepository
}

func NewService(repo repository.AttributeRepository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) CreateAttribute(ctx context.Context, req domain.Attribute) (int64, error) {
	return s.repo.CreateAttribute(ctx, req)
}

func (s *service) SearchAttributeFieldsByModelUid(ctx context.Context, modelUid string) ([]string, error) {
	return s.repo.SearchAttributeFieldsByModelUid(ctx, modelUid)
}

func (s *service) ListAttributes(ctx context.Context, modelUid string) ([]domain.Attribute, int64, error) {
	var (
		total int64
		attrs []domain.Attribute
		eg    errgroup.Group
	)
	eg.Go(func() error {
		var err error
		attrs, err = s.repo.ListAttributes(ctx, modelUid)
		return err
	})
	eg.Go(func() error {
		var err error
		total, err = s.repo.Total(ctx, modelUid)
		return err
	})
	if err := eg.Wait(); err != nil {
		return attrs, total, err
	}
	return attrs, total, nil
}

func (s *service) CustomAttributeFieldColumns(ctx *gin.Context, modelUid string, customField []string) (int64, error) {
	var (
		total int64
		eg    errgroup.Group
	)
	eg.Go(func() error {
		var err error
		total, err = s.repo.CustomAttributeFieldColumns(ctx, modelUid, customField)
		fmt.Print(err)
		return err
	})
	eg.Go(func() error {
		var err error
		total, err = s.repo.CustomAttributeFieldColumnsReverse(ctx, modelUid, customField)
		return err
	})
	if err := eg.Wait(); err != nil {
		return total, err
	}
	return total, nil
}
