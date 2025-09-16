package service

import (
	"context"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/resource/internal/domain"
	"github.com/Duke1616/ecmdb/pkg/cryptox"
)

type EncryptedSvc interface {
	Service
}

type EncryptedResourceService struct {
	Service
	attrSvc    attribute.Service
	encryptKey string
}

func NewEncryptedResourceService(inner Service, attrSvc attribute.Service, encryptKey string) EncryptedSvc {
	return &EncryptedResourceService{
		Service:    inner,
		attrSvc:    attrSvc,
		encryptKey: encryptKey,
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

func (s *EncryptedResourceService) FindSecureData(ctx context.Context, id int64, fieldUid string) (string, error) {
	// 先调用原始方法获取加密的数据
	encryptedData, err := s.Service.FindSecureData(ctx, id, fieldUid)
	if err != nil {
		return "", err
	}

	// 对加密的数据进行解密, 断言 string 因为目前只有字符串加密
	decryptedData, err := cryptox.DecryptAES[string](s.encryptKey, encryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt secure data: %w", err)
	}

	return decryptedData, nil
}

func (s *EncryptedResourceService) processEncryption(ctx context.Context, req domain.Resource) (domain.Resource, error) {
	secureFields, err := s.attrSvc.SearchAttributeFieldsBySecure(ctx, []string{req.ModelUID})
	if err != nil {
		return req, err
	}

	fields, ok := secureFields[req.ModelUID]
	if !ok || len(fields) == 0 {
		return req, nil
	}

	// 转为 map 加速 contains 判断
	secureMap := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		secureMap[f] = struct{}{}
	}

	encryptedData := make(map[string]interface{}, len(req.Data))
	for key, value := range req.Data {
		if _, needEncrypt := secureMap[key]; needEncrypt {
			encrypted, err := cryptox.EncryptAES(s.encryptKey, value)
			if err != nil {
				return req, fmt.Errorf("encrypt field %s failed: %w", key, err)
			}
			encryptedData[key] = encrypted
		} else {
			encryptedData[key] = value
		}
	}
	req.Data = encryptedData
	return req, nil
}
