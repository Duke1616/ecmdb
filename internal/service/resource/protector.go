package service

import (
	"context"
	"fmt"

	attribute "github.com/Duke1616/ecmdb/internal/service/attribute"
	"github.com/Duke1616/ecmdb/pkg/cryptox"
	"github.com/gotomicro/ego/core/elog"
)

// IResourceProtector 统管 ECMDB 资产在存储、运行和展示阶段的敏感属性保护
type IResourceProtector interface {
	// EncryptResource 根据模型敏感属性配置，对资产入库数据执行加密
	EncryptResource(ctx context.Context, modelUID string, data map[string]any) (map[string]any, error)

	// DecryptResource 严格解密资产中的敏感属性，用于内部运行时或需要明文凭据的场景
	DecryptResource(ctx context.Context, modelUID string, data map[string]any) (map[string]any, error)

	// MaskResource 将资产敏感字段替换为 "[已脱敏]"，用于列表展示、详情查看及审计日志
	MaskResource(ctx context.Context, modelUID string, data map[string]any) (map[string]any, error)

	// IsMasked 判断值是否为脱敏占位文本（用于防止编辑保存时将掩码误写入库）
	IsMasked(val any) bool
}

type resourceProtector struct {
	attrSvc   attribute.Service
	protector cryptox.ValueProtector
	logger    *elog.Component
}

// NewResourceProtector 创建资产敏感属性保护器
func NewResourceProtector(attrSvc attribute.Service, cipher cryptox.Crypto) IResourceProtector {
	return &resourceProtector{
		attrSvc:   attrSvc,
		protector: cryptox.NewValueProtector(cipher),
		logger:    elog.DefaultLogger,
	}
}

func (p *resourceProtector) IsMasked(val any) bool {
	str, ok := val.(string)
	return ok && str == cryptox.DefaultMask
}

func (p *resourceProtector) getSecureFields(ctx context.Context, modelUID string) ([]string, error) {
	if p.attrSvc == nil {
		return nil, nil
	}
	secureFieldsMap, err := p.attrSvc.SearchAttributeFieldsBySecure(ctx, []string{modelUID})
	if err != nil {
		return nil, fmt.Errorf("查询模型 %s 敏感属性失败: %w", modelUID, err)
	}
	return secureFieldsMap[modelUID], nil
}

// EncryptResource 遍历安全属性字段并执行加密
func (p *resourceProtector) EncryptResource(ctx context.Context, modelUID string, data map[string]any) (map[string]any, error) {
	if len(data) == 0 {
		return data, nil
	}

	secureFields, err := p.getSecureFields(ctx, modelUID)
	if err != nil {
		return nil, err
	}
	if len(secureFields) == 0 {
		return data, nil
	}

	result := make(map[string]any, len(data))
	for k, v := range data {
		result[k] = v
	}

	for _, field := range secureFields {
		val, exists := result[field]
		if !exists || val == nil {
			continue
		}

		strVal, ok := val.(string)
		if !ok || strVal == "" || p.IsMasked(strVal) {
			continue
		}

		encrypted, err := p.protector.Encrypt(strVal)
		if err != nil {
			p.logger.Error("资产属性加密失败", elog.String("model_uid", modelUID), elog.String("field", field), elog.FieldErr(err))
			return nil, fmt.Errorf("资产字段 %s 加密失败: %w", field, err)
		}
		result[field] = encrypted
	}

	return result, nil
}

// DecryptResource 遍历安全属性字段并执行解密（采用 Fail-Closed 严格解密）
func (p *resourceProtector) DecryptResource(ctx context.Context, modelUID string, data map[string]any) (map[string]any, error) {
	if len(data) == 0 {
		return data, nil
	}

	secureFields, err := p.getSecureFields(ctx, modelUID)
	if err != nil {
		return nil, err
	}
	if len(secureFields) == 0 {
		return data, nil
	}

	result := make(map[string]any, len(data))
	for k, v := range data {
		result[k] = v
	}

	for _, field := range secureFields {
		val, exists := result[field]
		if !exists || val == nil {
			continue
		}

		strVal, ok := val.(string)
		if !ok || strVal == "" {
			continue
		}

		decrypted, err := p.protector.DecryptCiphertext(strVal)
		if err != nil {
			// 若非合法密文，记录告警，回退至兼容解密
			p.logger.Warn("资产属性严格解密失败，尝试兼容解密", elog.String("model_uid", modelUID), elog.String("field", field), elog.FieldErr(err))
			decrypted, err = p.protector.Decrypt(strVal)
			if err != nil {
				p.logger.Error("资产属性解密失败", elog.String("model_uid", modelUID), elog.String("field", field), elog.FieldErr(err))
				return nil, fmt.Errorf("资产字段 %s 解密失败: %w", field, err)
			}
		}
		result[field] = decrypted
	}

	return result, nil
}

// MaskResource 将所有敏感属性字段值遮蔽为统一掩码占位符
func (p *resourceProtector) MaskResource(ctx context.Context, modelUID string, data map[string]any) (map[string]any, error) {
	if len(data) == 0 {
		return data, nil
	}

	secureFields, err := p.getSecureFields(ctx, modelUID)
	if err != nil {
		return nil, err
	}
	if len(secureFields) == 0 {
		return data, nil
	}

	result := make(map[string]any, len(data))
	for k, v := range data {
		result[k] = v
	}

	for _, field := range secureFields {
		val, exists := result[field]
		if !exists || val == nil {
			continue
		}

		strVal, ok := val.(string)
		if ok && strVal != "" {
			result[field] = cryptox.DefaultMask
		}
	}

	return result, nil
}
