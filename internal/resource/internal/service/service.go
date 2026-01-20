package service

import (
	"context"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/resource/internal/domain"
	"github.com/Duke1616/ecmdb/internal/resource/internal/repository"
	"golang.org/x/sync/errgroup"
)

//go:generate mockgen -source=./service.go -destination=../../mocks/resource.mock.go -package=resourcemocks -typed Service
type Service interface {
	// CreateResource 创建资产
	CreateResource(ctx context.Context, req domain.Resource) (int64, error)

	// BatchCreateOrUpdate 批量创建或修改资产
	// 基于 model_uid + name 进行 upsert,适用于 Excel 导入等批量操作
	BatchCreateOrUpdate(ctx context.Context, rs []domain.Resource) error

	// FindResourceById 根据ID，获取资产信息
	FindResourceById(ctx context.Context, fields []string, id int64) (domain.Resource, error)

	// ListResource 获取置顶模型资产数据
	ListResource(ctx context.Context, fields []string, modelUid string, offset, limit int64) ([]domain.Resource,
		int64, error)

	CountByModelUid(ctx context.Context, modelUid string) (int64, error)

	// SetCustomField 变更指定字段的数据
	SetCustomField(ctx context.Context, id int64, field string, data interface{}) (int64, error)

	// ListResourceByIds 资源关联关系调用，查询关联数据
	ListResourceByIds(ctx context.Context, fields []string, ids []int64) ([]domain.Resource, error)

	// ListExcludeAndFilterResourceByIds 排序以及过滤
	ListExcludeAndFilterResourceByIds(ctx context.Context, fields []string, modelUid string, offset, limit int64,
		ids []int64, filter domain.Condition) ([]domain.Resource, int64, error)

	// DeleteResource 删除资产数据
	DeleteResource(ctx context.Context, id int64) (int64, error)

	// CountByModelUids 聚合查看模型下的数量
	CountByModelUids(ctx context.Context, modelUids []string) (map[string]int, error)

	// Search 全局搜索
	Search(ctx context.Context, text string) ([]domain.SearchResource, error)

	// FindSecureData 查看指定资产加密字段数据
	FindSecureData(ctx context.Context, id int64, fieldUid string) (string, error)

	// UpdateResource 修改资产数据
	UpdateResource(ctx context.Context, resource domain.Resource) (int64, error)

	// BatchUpdateResources 因为资产属性变更，处理改变
	BatchUpdateResources(ctx context.Context, resources []domain.Resource) (int64, error)

	// ListBeforeUtime 获取指定时间前的资产列表
	ListBeforeUtime(ctx context.Context, utime int64, fields []string, modelUid string,
		offset, limit int64) ([]domain.Resource, error)
}

type service struct {
	repo repository.ResourceRepository
}

func (s *service) BatchCreateOrUpdate(ctx context.Context, rs []domain.Resource) error {
	return s.repo.BatchCreateOrUpdate(ctx, rs)
}

func (s *service) ListBeforeUtime(ctx context.Context, utime int64, fields []string, modelUid string,
	offset, limit int64) ([]domain.Resource, error) {
	return s.repo.ListBeforeUtime(ctx, utime, fields, modelUid, offset, limit)
}

func (s *service) BatchUpdateResources(ctx context.Context, resources []domain.Resource) (int64, error) {
	return s.repo.BatchUpdateResources(ctx, resources)
}

func (s *service) SetCustomField(ctx context.Context, id int64, field string, data interface{}) (int64, error) {
	return s.repo.SetCustomField(ctx, id, field, data)
}

func (s *service) CountByModelUid(ctx context.Context, modelUid string) (int64, error) {
	return s.repo.TotalByModelUid(ctx, modelUid)
}

func NewService(repo repository.ResourceRepository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) UpdateResource(ctx context.Context, resource domain.Resource) (int64, error) {
	return s.repo.UpdateResource(ctx, resource)
}

func (s *service) CreateResource(ctx context.Context, req domain.Resource) (int64, error) {
	return s.repo.CreateResource(ctx, req)
}

func (s *service) FindResourceById(ctx context.Context, fields []string, id int64) (domain.Resource, error) {
	return s.repo.FindResourceById(ctx, fields, id)
}

func (s *service) ListResource(ctx context.Context, fields []string, modelUid string, offset, limit int64) ([]domain.Resource, int64, error) {
	if fields == nil {
		return nil, 0, fmt.Errorf("传递字段信息不能为空")
	}

	if modelUid == "" {
		return nil, 0, fmt.Errorf("模型唯一标识不能为空")
	}

	var (
		total     int64
		resources []domain.Resource
		eg        errgroup.Group
	)
	eg.Go(func() error {
		var err error
		resources, err = s.repo.ListResource(ctx, fields, modelUid, offset, limit)
		return err
	})
	eg.Go(func() error {
		var err error
		total, err = s.repo.TotalByModelUid(ctx, modelUid)
		return err
	})
	if err := eg.Wait(); err != nil {
		return resources, total, err
	}
	return resources, total, nil
}

func (s *service) ListResourceByIds(ctx context.Context, fields []string, ids []int64) ([]domain.Resource, error) {
	return s.repo.ListResourcesByIds(ctx, fields, ids)
}

func (s *service) ListExcludeAndFilterResourceByIds(ctx context.Context, fields []string, modelUid string, offset,
	limit int64, ids []int64, filter domain.Condition) ([]domain.Resource, int64, error) {
	var (
		total     int64
		resources []domain.Resource
		eg        errgroup.Group
	)
	eg.Go(func() error {
		var err error
		resources, err = s.repo.ListExcludeAndFilterResourceByIds(ctx, fields, modelUid, offset, limit, ids, filter)
		return err
	})
	eg.Go(func() error {
		var err error
		total, err = s.repo.TotalExcludeAndFilterResourceByIds(ctx, modelUid, ids, filter)
		return err
	})
	if err := eg.Wait(); err != nil {
		return resources, total, err
	}
	return resources, total, nil
}

func (s *service) DeleteResource(ctx context.Context, id int64) (int64, error) {
	return s.repo.DeleteResource(ctx, id)
}

func (s *service) CountByModelUids(ctx context.Context, modelUids []string) (map[string]int, error) {
	return s.repo.CountByModelUids(ctx, modelUids)
}

func (s *service) Search(ctx context.Context, text string) ([]domain.SearchResource, error) {
	return s.repo.Search(ctx, text)
}

func (s *service) FindSecureData(ctx context.Context, id int64, fieldUid string) (string, error) {
	return s.repo.FindSecureData(ctx, id, fieldUid)
}

func (s *service) desensitization(ctx context.Context, modelUid string) ([]domain.Resource, error) {
	return nil, nil
}
