package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/resource/internal/domain"
	"github.com/Duke1616/ecmdb/pkg/cryptox"
	"github.com/ecodeclub/ekit/slice"
)

type EncryptedSvc interface {
	Service
}

type EncryptedResourceService struct {
	Service
	attrSvc attribute.Service
	crypto  cryptox.Crypto[string]
}

func NewEncryptedResourceService(inner Service, attrSvc attribute.Service, aseKey string) EncryptedSvc {
	return &EncryptedResourceService{
		Service: inner,
		attrSvc: attrSvc,
		crypto:  cryptox.NewAESCrypto[string](aseKey),
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

// decryptResources 批量解密资源数据
func (s *EncryptedResourceService) decryptResources(ctx context.Context, resources []domain.Resource) ([]domain.Resource, error) {
	// 收集所有模型UID
	uniqueMap := &sync.Map{}
	modelUIDs := slice.FilterMap(resources, func(idx int, src domain.Resource) (string, bool) {
		_, loaded := uniqueMap.LoadOrStore(src.ModelUID, struct{}{})
		return src.ModelUID, !loaded
	})

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

// encryptValue 加密值（目前只支持字符串类型）
func (s *EncryptedResourceService) encryptValue(value interface{}) (interface{}, error) {
	strVal, ok := value.(string)
	if !ok {
		// 非字符串类型直接返回原值
		return value, nil
	}

	return s.crypto.Encrypt(strVal)
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

// decryptValue 解密值（目前只支持字符串类型）
func (s *EncryptedResourceService) decryptValue(value interface{}) (interface{}, error) {
	strVal, ok := value.(string)
	if !ok {
		// 非字符串类型直接返回原值
		return value, nil
	}

	return s.crypto.Decrypt(strVal)
}

// encode 编码敏感字段（已废弃，使用 encryptSensitiveFields 替代）
func (s *EncryptedResourceService) encode(data map[string]interface{}, secureFields []string) (map[string]interface{}, error) {
	return s.encryptSensitiveFields(data, secureFields)
}

// decode 解码敏感字段（已废弃，使用 decryptSensitiveFields 替代）
func (s *EncryptedResourceService) decode(data map[string]interface{}, secureFields []string) (map[string]interface{}, error) {
	return s.decryptSensitiveFields(data, secureFields)
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
		return "", fmt.Errorf("failed to decrypt secure data for field %s: %w", fieldUid, err)
	}

	return decryptedData, nil
}
