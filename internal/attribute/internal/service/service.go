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
	// CreateAttribute 创建模型字段
	CreateAttribute(ctx context.Context, req domain.Attribute) (int64, error)

	// SearchAttributeFieldsByModelUid 查询模型下的所有字段信息，不包含安全字段，内部使用
	SearchAttributeFieldsByModelUid(ctx context.Context, modelUid string) ([]string, error)

	// SearchAllAttributeFieldsByModelUid 查询模型下的所有字段信息，内部使用
	SearchAllAttributeFieldsByModelUid(ctx context.Context, modelUid string) ([]string, error)

	// SearchAttributeFieldsBySecure 查询全有的安全字段
	SearchAttributeFieldsBySecure(ctx context.Context, modelUids []string) (map[string][]string, error)

	// ListAttributes 查询模型下的所有字段详情信息，前端使用
	ListAttributes(ctx context.Context, modelUID string) ([]domain.Attribute, int64, error)

	// DeleteAttribute 删除模型字段
	DeleteAttribute(ctx context.Context, id int64) (int64, error)

	// UpdateAttribute 更新模型字段
	UpdateAttribute(ctx context.Context, attribute domain.Attribute) (int64, error)

	// CustomAttributeFieldColumns 自定义展示字段、以及排序
	CustomAttributeFieldColumns(ctx *gin.Context, modelUid string, customField []string) (int64, error)

	// ListAttributePipeline 根据组聚合获取每个组下的所有字段
	ListAttributePipeline(ctx *gin.Context, modelUid string) ([]domain.AttributePipeline, error)

	// CreateDefaultAttribute 创建新模型，创建默认字段信息
	CreateDefaultAttribute(ctx context.Context, modelUid string) (int64, error)

	// CreateAttributeGroup 创建模型字段组
	CreateAttributeGroup(ctx context.Context, req domain.AttributeGroup) (int64, error)

	// ListAttributeGroup 模型组
	ListAttributeGroup(ctx context.Context, modelUid string) ([]domain.AttributeGroup, error)

	// ListAttributeGroupByIds 返回每个组下面的属性
	ListAttributeGroupByIds(ctx context.Context, ids []int64) ([]domain.AttributeGroup, error)
}

type service struct {
	repo      repository.AttributeRepository
	groupRepo repository.AttributeGroupRepository
}

func (s *service) UpdateAttribute(ctx context.Context, attribute domain.Attribute) (int64, error) {
	return s.repo.UpdateAttribute(ctx, attribute)
}

func NewService(repo repository.AttributeRepository, groupRepo repository.AttributeGroupRepository) Service {
	return &service{
		repo:      repo,
		groupRepo: groupRepo,
	}
}

func (s *service) SearchAllAttributeFieldsByModelUid(ctx context.Context, modelUid string) ([]string, error) {
	return s.repo.SearchAllAttributeFieldsByModelUid(ctx, modelUid)
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

func (s *service) DeleteAttribute(ctx context.Context, id int64) (int64, error) {
	return s.repo.DeleteAttribute(ctx, id)
}

func (s *service) CreateDefaultAttribute(ctx context.Context, modelUid string) (int64, error) {
	groupId, err := s.CreateAttributeGroup(ctx, domain.AttributeGroup{
		Name:     "基础属性",
		ModelUid: modelUid,
		Index:    0,
	})
	if err != nil {
		return 0, err
	}

	fmt.Println("获取ID", groupId)

	attr := s.defaultAttr(modelUid, groupId)

	return s.repo.CreateAttribute(ctx, attr)
}

func (s *service) ListAttributeGroupByIds(ctx context.Context, ids []int64) ([]domain.AttributeGroup, error) {
	return s.groupRepo.ListAttributeGroupByIds(ctx, ids)
}

func (s *service) ListAttributeGroup(ctx context.Context, modelUid string) ([]domain.AttributeGroup, error) {
	return s.groupRepo.ListAttributeGroup(ctx, modelUid)
}

func (s *service) CreateAttributeGroup(ctx context.Context, req domain.AttributeGroup) (int64, error) {
	return s.groupRepo.CreateAttributeGroup(ctx, req)
}

func (s *service) ListAttributePipeline(ctx *gin.Context, modelUid string) ([]domain.AttributePipeline, error) {
	return s.repo.ListAttributePipeline(ctx, modelUid)
}

func (s *service) SearchAttributeFieldsBySecure(ctx context.Context, modelUids []string) (map[string][]string, error) {
	return s.repo.SearchAttributeFieldsBySecure(ctx, modelUids)
}

func (s *service) defaultAttr(modelUid string, groupId int64) domain.Attribute {
	return domain.Attribute{
		ModelUid:  modelUid,
		Index:     0,
		Display:   true,
		Required:  true,
		FieldName: "名称",
		FieldType: "string",
		FieldUid:  "name",
		GroupId:   groupId,
		Secure:    false,
	}
}
