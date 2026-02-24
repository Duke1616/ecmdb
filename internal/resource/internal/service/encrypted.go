package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/resource/internal/domain"
	"github.com/Duke1616/ecmdb/pkg/cryptox"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
)

//go:generate mockgen -source=./encrypted.go -destination=../../mocks/encrypted.mock.go -package=resourcemocks -typed EncryptedSvc
type EncryptedSvc interface {
	Service
	// ListAndDecryptBeforeUtime 列出指定时间前的资源并解密指定字段
	ListAndDecryptBeforeUtime(ctx context.Context, utime int64, fields []string, modelUid string,
		offset, limit int64) ([]domain.Resource, error)
}

type EncryptedResourceService struct {
	Service
	attrSvc attribute.Service
	crypto  cryptox.Crypto
	logger  *elog.Component
}

func (s *EncryptedResourceService) ListAndDecryptBeforeUtime(ctx context.Context, utime int64, fields []string, modelUid string, offset, limit int64) ([]domain.Resource, error) {
	resources, err := s.Service.ListBeforeUtime(ctx, utime, fields, modelUid, offset, limit)

	// 无论是否需要解密都进行操作
	for i := range resources {
		decryptedData, err1 := s.decryptSensitiveFields(resources[i].Data, fields)
		if err1 != nil {
			return nil, fmt.Errorf("failed to decrypt resource %d: %w", resources[i].ID, err)
		}

		resources[i].Data = decryptedData
	}

	return resources, nil
}

func NewEncryptedResourceService(inner Service, attrSvc attribute.Service, crypto cryptox.Crypto) EncryptedSvc {
	return &EncryptedResourceService{
		Service: inner,
		attrSvc: attrSvc,
		crypto:  crypto,
		logger:  elog.DefaultLogger,
	}
}

func (s *EncryptedResourceService) CreateResource(ctx context.Context, req domain.Resource) (int64, error) {
	// 加密处理
	encryptedReq, err := s.processEncryption(ctx, req)
	if err != nil {
		return 0, err
	}

	return s.Service.CreateResource(ctx, encryptedReq)
}

func (s *EncryptedResourceService) ListResource(ctx context.Context, fields []string, modelUid string, offset, limit int64) (
	[]domain.Resource,
	int64, error) {
	rs, total, err := s.Service.ListResource(ctx, fields, modelUid, offset, limit)

	if err != nil {
		return nil, 0, err
	}

	if len(rs) == 0 {
		return nil, 0, err
	}

	// 批量解密资源数据
	decodeRs, err := s.decryptResources(ctx, rs)

	return decodeRs, total, err
}

func (s *EncryptedResourceService) BatchUpdateResources(ctx context.Context, resources []domain.Resource) (int64, error) {
	if len(resources) == 0 {
		return 0, nil
	}

	// 收集所有模型UID
	modelUIDs := s.buildModelUIDs(resources)

	// 批量查询安全字段
	secureFieldsMap, err := s.attrSvc.SearchAttributeFieldsBySecure(ctx, modelUIDs)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch secure fields: %w", err)
	}

	// 处理如果需要加密则进行加密处理
	for i := range resources {
		secureFields := secureFieldsMap[resources[i].ModelUID]
		if len(secureFields) == 0 {
			continue
		}

		encryptedData, err1 := s.encryptSensitiveFields(resources[i].Data, secureFieldsMap[resources[i].ModelUID])
		if err1 != nil {
			return 0, fmt.Errorf("failed to encrypt sensitive fields: %w", err)
		}

		// 更新处理后的值
		resources[i].Data = encryptedData
	}

	return s.Service.BatchUpdateResources(ctx, resources)
}

func (s *EncryptedResourceService) UpdateResource(ctx context.Context, req domain.Resource) (int64, error) {
	encryptedReq, err := s.processEncryption(ctx, req)
	if err != nil {
		return 0, err

	}

	return s.Service.UpdateResource(ctx, encryptedReq)
}

func (s *EncryptedResourceService) ListResourceByIds(ctx context.Context, fields []string, ids []int64) ([]domain.Resource, error) {
	rs, err := s.Service.ListResourceByIds(ctx, fields, ids)
	if err != nil {
		return nil, err
	}

	if len(rs) == 0 {
		return rs, nil
	}

	// 批量解密资源数据
	return s.decryptResources(ctx, rs)
}

func (s *EncryptedResourceService) FindResourceById(ctx context.Context, fields []string, id int64) (domain.Resource, error) {
	// 获取资源
	resource, err := s.Service.FindResourceById(ctx, fields, id)
	if err != nil {
		return resource, err
	}

	// 获取需要解密的安全字段
	secureFields, err := s.getSecureFields(ctx, resource.ModelUID)
	if err != nil {
		return resource, fmt.Errorf("failed to get secure fields: %w", err)
	}

	if len(secureFields) == 0 {
		return resource, nil
	}

	// 解密敏感字段
	decryptedData, err := s.decryptSensitiveFields(resource.Data, secureFields)
	if err != nil {
		return resource, fmt.Errorf("failed to decrypt resource %d: %w", resource.ID, err)
	}

	resource.Data = decryptedData

	return resource, nil
}

func (s *EncryptedResourceService) buildModelUIDs(resources []domain.Resource) []string {
	uniqueMap := &sync.Map{}
	return slice.FilterMap(resources, func(idx int, src domain.Resource) (string, bool) {
		_, loaded := uniqueMap.LoadOrStore(src.ModelUID, struct{}{})
		return src.ModelUID, !loaded
	})
}

// decryptResources 批量解密资源数据
func (s *EncryptedResourceService) decryptResources(ctx context.Context, resources []domain.Resource) ([]domain.Resource, error) {
	// 收集所有模型UID
	modelUIDs := s.buildModelUIDs(resources)

	// 批量查询安全字段
	secureFieldsMap, err := s.attrSvc.SearchAttributeFieldsBySecure(ctx, modelUIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch secure fields: %w", err)
	}

	// 解密每个资源的数据
	for i := range resources {
		secureFields := secureFieldsMap[resources[i].ModelUID]
		if len(secureFields) == 0 {
			continue
		}

		decryptedData, err1 := s.decryptSensitiveFields(resources[i].Data, secureFields)
		if err1 != nil {
			return nil, fmt.Errorf("failed to decrypt resource %d: %w", resources[i].ID, err)
		}

		resources[i].Data = decryptedData
	}

	return resources, nil
}

func (s *EncryptedResourceService) processEncryption(ctx context.Context, req domain.Resource) (domain.Resource, error) {
	// 获取需要加密的字段
	secureFields, err := s.getSecureFields(ctx, req.ModelUID)
	if err != nil {
		return req, fmt.Errorf("failed to get secure fields: %w", err)
	}

	if len(secureFields) == 0 {
		return req, nil
	}

	// 加密敏感字段
	encryptedData, err := s.encryptSensitiveFields(req.Data, secureFields)
	if err != nil {
		return req, fmt.Errorf("failed to encrypt sensitive fields: %w", err)
	}

	req.Data = encryptedData
	return req, nil
}

// getSecureFields 获取需要加密的字段列表
func (s *EncryptedResourceService) getSecureFields(ctx context.Context, modelUID string) ([]string, error) {
	secureFieldsMap, err := s.attrSvc.SearchAttributeFieldsBySecure(ctx, []string{modelUID})
	if err != nil {
		return nil, err
	}
	return secureFieldsMap[modelUID], nil
}

// encryptSensitiveFields 加密敏感字段
func (s *EncryptedResourceService) encryptSensitiveFields(data map[string]interface{}, secureFields []string) (map[string]interface{}, error) {
	// 构建安全字段映射，提高查找效率
	secureMap := s.buildSecureFieldMap(secureFields)

	result := make(map[string]interface{}, len(data))

	for key, value := range data {
		if s.isSecureField(key, secureMap) {
			encryptedValue, err := s.encryptValue(value)
			if err != nil {
				return nil, fmt.Errorf("encrypt field %s failed: %w", key, err)
			}
			result[key] = encryptedValue
		} else {
			result[key] = value
		}
	}

	return result, nil
}

// buildSecureFieldMap 构建安全字段映射
func (s *EncryptedResourceService) buildSecureFieldMap(secureFields []string) map[string]struct{} {
	secureMap := make(map[string]struct{}, len(secureFields))
	for _, field := range secureFields {
		secureMap[field] = struct{}{}
	}
	return secureMap
}

// isSecureField 判断字段是否为安全字段
func (s *EncryptedResourceService) isSecureField(fieldName string, secureMap map[string]struct{}) bool {
	_, exists := secureMap[fieldName]
	return exists
}

// decryptSensitiveFields 解密敏感字段
func (s *EncryptedResourceService) decryptSensitiveFields(data map[string]interface{}, secureFields []string) (map[string]interface{}, error) {
	// 构建安全字段映射，提高查找效率
	secureMap := s.buildSecureFieldMap(secureFields)

	result := make(map[string]interface{}, len(data))

	for key, value := range data {
		if s.isSecureField(key, secureMap) {
			decryptedValue, err := s.decryptValue(value)
			if err != nil {
				return nil, fmt.Errorf("decrypt field %s failed: %w", key, err)
			}
			result[key] = decryptedValue
		} else {
			result[key] = value
		}
	}

	return result, nil
}

// encryptValue 加密值
func (s *EncryptedResourceService) encryptValue(value interface{}) (interface{}, error) {
	strVal, ok := value.(string)
	if !ok {
		return value, nil
	}

	val, err := s.crypto.Encrypt(strVal)
	if err != nil {
		s.logger.Error("encrypt failed", elog.FieldErr(err), elog.FieldValue(strVal))
	}
	return val, nil
}

// decryptValue 解密值
func (s *EncryptedResourceService) decryptValue(value interface{}) (interface{}, error) {
	strVal, ok := value.(string)
	if !ok {
		return value, nil
	}

	val, err := s.crypto.Decrypt(strVal)
	if err != nil {
		s.logger.Error("decrypt failed", elog.FieldErr(err), elog.FieldValue(strVal))
	}
	return val, nil
}

func (s *EncryptedResourceService) FindSecureData(ctx context.Context, id int64, fieldUid string) (string, error) {
	// 获取加密的数据
	encryptedData, err := s.Service.FindSecureData(ctx, id, fieldUid)
	if err != nil {
		return "", fmt.Errorf("failed to get secure data: %w", err)
	}

	// 解密数据
	decryptedData, err := s.crypto.Decrypt(encryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt secure data: %w", err)
	}

	return decryptedData, nil
}

// BatchCreateOrUpdate 批量创建或更新资产,支持加密字段
func (s *EncryptedResourceService) BatchCreateOrUpdate(ctx context.Context, resources []domain.Resource) error {
	if len(resources) == 0 {
		return nil
	}

	// 收集所有模型UID
	modelUIDs := s.buildModelUIDs(resources)

	// 批量查询安全字段
	secureFieldsMap, err := s.attrSvc.SearchAttributeFieldsBySecure(ctx, modelUIDs)
	if err != nil {
		return fmt.Errorf("failed to fetch secure fields: %w", err)
	}

	// 处理如果需要加密则进行加密处理
	for i := range resources {
		secureFields := secureFieldsMap[resources[i].ModelUID]
		if len(secureFields) == 0 {
			continue
		}

		encryptedData, err1 := s.encryptSensitiveFields(resources[i].Data, secureFieldsMap[resources[i].ModelUID])
		if err1 != nil {
			return fmt.Errorf("failed to encrypt sensitive fields: %w", err1)
		}

		// 更新处理后的值
		resources[i].Data = encryptedData
	}

	return s.Service.BatchCreateOrUpdate(ctx, resources)
}

func (s *EncryptedResourceService) ListResourcesWithFilters(ctx context.Context, fields []string, modelUid string, ids []int64, offset, limit int64,
	filterGroups []domain.FilterGroup) ([]domain.Resource, int64, error) {
	rs, total, err := s.Service.ListResourcesWithFilters(ctx, fields, modelUid, ids, offset, limit, filterGroups)
	if err != nil {
		return nil, 0, err
	}

	if len(rs) == 0 {
		return rs, total, nil
	}

	// 批量解密资源数据
	decodedRs, err := s.decryptResources(ctx, rs)
	return decodedRs, total, err
}
