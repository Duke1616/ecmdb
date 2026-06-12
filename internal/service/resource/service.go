package service

import (
	"context"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/domain"
	"github.com/Duke1616/ecmdb/internal/repository"
	attribute "github.com/Duke1616/ecmdb/internal/service/attribute"
	"github.com/Duke1616/ecmdb/pkg/cryptox"
	"github.com/gotomicro/ego/core/elog"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

type EncryptedSvc = Service

//go:generate mockgen -source=./service.go -destination=../../mocks/resource.mock.go -package=resourcemocks -typed Service
type Service interface {
	// CreateResource 创建资产
	CreateResource(ctx context.Context, req domain.Resource) (int64, error)

	// BatchCreateOrUpdate 批量创建或修改资产
	BatchCreateOrUpdate(ctx context.Context, rs []domain.Resource) error

	// FindResourceById 根据ID，获取资产信息
	FindResourceById(ctx context.Context, fields []string, id int64) (domain.Resource, error)

	// ListResource 获取资产数据
	ListResource(ctx context.Context, fields []string, modelUid string, offset, limit int64) ([]domain.Resource,
		int64, error)

	CountByModelUid(ctx context.Context, modelUid string) (int64, error)

	// SetCustomField 变更指定字段的数据
	SetCustomField(ctx context.Context, id int64, field string, data interface{}) (int64, error)

	// UnsetCustomField 抹除指定模型下所有资产的自定义字段
	UnsetCustomField(ctx context.Context, modelUid string, fieldUid string) (int64, error)

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

	// ListAndDecryptBeforeUtime 获取指定时间前的资产列表并解密
	ListAndDecryptBeforeUtime(ctx context.Context, utime int64, fields []string, modelUid string,
		offset, limit int64) ([]domain.Resource, error)

	// ListResourcesWithFilters 根据复杂筛选条件获取资产列表
	ListResourcesWithFilters(ctx context.Context, fields []string, modelUid string, ids []int64, offset, limit int64,
		filterGroups []domain.FilterGroup) ([]domain.Resource, int64, error)

	// CheckBeforeDelete 检查指定模型下是否还有资产实例
	CheckBeforeDelete(ctx context.Context, modelUid string) error
}

type service struct {
	repo    repository.ResourceRepository
	attrSvc attribute.Service
	crypto  cryptox.Crypto
	logger  *elog.Component
}

func NewService(repo repository.ResourceRepository, attrSvc attribute.Service, crypto cryptox.Crypto) Service {
	return &service{
		repo:    repo,
		attrSvc: attrSvc,
		crypto:  crypto,
		logger:  elog.DefaultLogger,
	}
}

func (s *service) CreateResource(ctx context.Context, req domain.Resource) (int64, error) {
	encryptedReq, err := s.encryptResource(ctx, req)
	if err != nil {
		return 0, err
	}
	return s.repo.CreateResource(ctx, encryptedReq)
}

func (s *service) UpdateResource(ctx context.Context, req domain.Resource) (int64, error) {
	encryptedReq, err := s.encryptResource(ctx, req)
	if err != nil {
		return 0, err
	}
	return s.repo.UpdateResource(ctx, encryptedReq)
}

func (s *service) BatchCreateOrUpdate(ctx context.Context, resources []domain.Resource) error {
	encryptedRs, err := s.encryptResources(ctx, resources)
	if err != nil {
		return err
	}
	return s.repo.BatchCreateOrUpdate(ctx, encryptedRs)
}

func (s *service) BatchUpdateResources(ctx context.Context, resources []domain.Resource) (int64, error) {
	encryptedRs, err := s.encryptResources(ctx, resources)
	if err != nil {
		return 0, err
	}
	return s.repo.BatchUpdateResources(ctx, encryptedRs)
}

func (s *service) FindResourceById(ctx context.Context, fields []string, id int64) (domain.Resource, error) {
	resource, err := s.repo.FindResourceById(ctx, fields, id)
	if err != nil {
		return resource, err
	}

	secureFields, err := s.getSecureFields(ctx, resource.ModelUID)
	if err != nil {
		return resource, fmt.Errorf("failed to get secure fields: %w", err)
	}

	if len(secureFields) == 0 {
		return resource, nil
	}

	decryptedData, err := s.decryptSensitiveFields(resource.Data, secureFields)
	if err != nil {
		return resource, fmt.Errorf("failed to decrypt resource %d: %w", resource.ID, err)
	}

	resource.Data = decryptedData
	return resource, nil
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

	if len(resources) == 0 {
		return resources, total, nil
	}

	decodedRs, err := s.decryptResources(ctx, resources)
	return decodedRs, total, err
}

func (s *service) ListResourceByIds(ctx context.Context, fields []string, ids []int64) ([]domain.Resource, error) {
	rs, err := s.repo.ListResourcesByIds(ctx, fields, ids)
	if err != nil {
		return nil, err
	}

	if len(rs) == 0 {
		return rs, nil
	}

	return s.decryptResources(ctx, rs)
}

func (s *service) ListResourcesWithFilters(ctx context.Context, fields []string, modelUid string, ids []int64, offset, limit int64,
	filterGroups []domain.FilterGroup) ([]domain.Resource, int64, error) {
	var (
		total     int64
		resources []domain.Resource
		eg        errgroup.Group
	)

	eg.Go(func() error {
		var err error
		resources, err = s.repo.ListResourcesWithFilters(ctx, fields, modelUid, ids, offset, limit, filterGroups)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.TotalResourcesWithFilters(ctx, modelUid, ids, filterGroups)
		return err
	})

	if err := eg.Wait(); err != nil {
		return resources, total, err
	}

	if len(resources) == 0 {
		return resources, total, nil
	}

	decodedRs, err := s.decryptResources(ctx, resources)
	return decodedRs, total, err
}

func (s *service) ListAndDecryptBeforeUtime(ctx context.Context, utime int64, fields []string, modelUid string, offset, limit int64) ([]domain.Resource, error) {
	resources, err := s.repo.ListBeforeUtime(ctx, utime, fields, modelUid, offset, limit)
	if err != nil {
		return nil, err
	}

	for i := range resources {
		decryptedData, err1 := s.decryptSensitiveFields(resources[i].Data, fields)
		if err1 != nil {
			return nil, fmt.Errorf("failed to decrypt resource %d: %w", resources[i].ID, err1)
		}
		resources[i].Data = decryptedData
	}

	return resources, nil
}

func (s *service) FindSecureData(ctx context.Context, id int64, fieldUid string) (string, error) {
	encryptedData, err := s.repo.FindSecureData(ctx, id, fieldUid)
	if err != nil {
		return "", fmt.Errorf("failed to get secure data: %w", err)
	}

	decryptedData, err := s.crypto.Decrypt(encryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt secure data: %w", err)
	}

	return decryptedData, nil
}

func (s *service) ListBeforeUtime(ctx context.Context, utime int64, fields []string, modelUid string, offset, limit int64) ([]domain.Resource, error) {
	return s.repo.ListBeforeUtime(ctx, utime, fields, modelUid, offset, limit)
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

func (s *service) SetCustomField(ctx context.Context, id int64, field string, data interface{}) (int64, error) {
	return s.repo.SetCustomField(ctx, id, field, data)
}

func (s *service) UnsetCustomField(ctx context.Context, modelUid string, fieldUid string) (int64, error) {
	return s.repo.UnsetCustomField(ctx, modelUid, fieldUid)
}

func (s *service) CountByModelUid(ctx context.Context, modelUid string) (int64, error) {
	return s.repo.TotalByModelUid(ctx, modelUid)
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

func (s *service) CheckBeforeDelete(ctx context.Context, modelUid string) error {
	count, err := s.repo.TotalByModelUid(ctx, modelUid)
	if err != nil {
		return fmt.Errorf("资产检查异常: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("资产实例依赖拦截：模型 [%s] 正在被 %d 个资产实例使用，请先清空资产数据", modelUid, count)
	}
	return nil
}

// 辅助加解密实现

func (s *service) buildModelUIDs(resources []domain.Resource) []string {
	return lo.Uniq(lo.Map(resources, func(src domain.Resource, _ int) string {
		return src.ModelUID
	}))
}

// transformSensitiveFields 统一处理敏感字段的值转换逻辑 (加密/解密)
func (s *service) transformSensitiveFields(
	data map[string]interface{},
	secureFields []string,
	transform func(interface{}) (interface{}, error),
) (map[string]interface{}, error) {
	if len(secureFields) == 0 || len(data) == 0 {
		return data, nil
	}

	result := make(map[string]interface{}, len(data))
	for key, value := range data {
		if lo.Contains(secureFields, key) {
			transformedValue, err := transform(value)
			if err != nil {
				return nil, fmt.Errorf("transform field %s failed: %w", key, err)
			}
			result[key] = transformedValue
		} else {
			result[key] = value
		}
	}

	return result, nil
}

func (s *service) encryptSensitiveFields(data map[string]interface{}, secureFields []string) (map[string]interface{}, error) {
	return s.transformSensitiveFields(data, secureFields, s.encryptValue)
}

func (s *service) decryptSensitiveFields(data map[string]interface{}, secureFields []string) (map[string]interface{}, error) {
	return s.transformSensitiveFields(data, secureFields, s.decryptValue)
}

// transformResources 统一批量处理资源的加解密逻辑
func (s *service) transformResources(
	ctx context.Context,
	resources []domain.Resource,
	transform func(map[string]interface{}, []string) (map[string]interface{}, error),
) ([]domain.Resource, error) {
	if len(resources) == 0 {
		return resources, nil
	}

	modelUIDs := s.buildModelUIDs(resources)
	secureFieldsMap, err := s.attrSvc.SearchAttributeFieldsBySecure(ctx, modelUIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch secure fields: %w", err)
	}

	for i := range resources {
		secureFields := secureFieldsMap[resources[i].ModelUID]
		if len(secureFields) == 0 {
			continue
		}

		transformedData, err1 := transform(resources[i].Data, secureFields)
		if err1 != nil {
			return nil, err1
		}
		resources[i].Data = transformedData
	}

	return resources, nil
}

func (s *service) encryptResources(ctx context.Context, resources []domain.Resource) ([]domain.Resource, error) {
	return s.transformResources(ctx, resources, s.encryptSensitiveFields)
}

func (s *service) decryptResources(ctx context.Context, resources []domain.Resource) ([]domain.Resource, error) {
	return s.transformResources(ctx, resources, s.decryptSensitiveFields)
}

func (s *service) encryptResource(ctx context.Context, req domain.Resource) (domain.Resource, error) {
	secureFields, err := s.getSecureFields(ctx, req.ModelUID)
	if err != nil {
		return req, fmt.Errorf("failed to get secure fields: %w", err)
	}

	if len(secureFields) == 0 {
		return req, nil
	}

	encryptedData, err := s.encryptSensitiveFields(req.Data, secureFields)
	if err != nil {
		return req, fmt.Errorf("failed to encrypt sensitive fields: %w", err)
	}

	req.Data = encryptedData
	return req, nil
}

func (s *service) getSecureFields(ctx context.Context, modelUID string) ([]string, error) {
	secureFieldsMap, err := s.attrSvc.SearchAttributeFieldsBySecure(ctx, []string{modelUID})
	if err != nil {
		return nil, err
	}
	return secureFieldsMap[modelUID], nil
}

func (s *service) encryptValue(value interface{}) (interface{}, error) {
	strVal, ok := value.(string)
	if !ok {
		return value, nil
	}

	val, err := s.crypto.Encrypt(strVal)
	if err != nil {
		s.logger.Error("encrypt failed", elog.FieldErr(err), elog.FieldValue(strVal))
		return val, err
	}
	return val, nil
}

func (s *service) decryptValue(value interface{}) (interface{}, error) {
	strVal, ok := value.(string)
	if !ok {
		return value, nil
	}

	val, err := s.crypto.Decrypt(strVal)
	if err != nil {
		s.logger.Error("decrypt failed", elog.FieldErr(err), elog.FieldValue(strVal))
		return val, err
	}
	return val, nil
}
