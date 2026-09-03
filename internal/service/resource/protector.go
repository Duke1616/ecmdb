package service

import (
	"context"
	"fmt"
	"maps"
	"strings"

	"github.com/Duke1616/ecmdb/internal/domain"
	attribute "github.com/Duke1616/ecmdb/internal/service/attribute"
	"github.com/Duke1616/ecmdb/pkg/cryptox"
	"github.com/gotomicro/ego/core/elog"
	"github.com/samber/lo"
)

// IResourceProtector 统管 ECMDB 资产在存储、运行和展示阶段的敏感属性保护
type IResourceProtector interface {
	// Encrypt 对单个资产的敏感属性执行加密
	Encrypt(ctx context.Context, res domain.Resource) (domain.Resource, error)

	// EncryptMany 批量对资产执行敏感属性加密（聚合查询属性以优化性能）
	EncryptMany(ctx context.Context, resources []domain.Resource) ([]domain.Resource, error)

	// Decrypt 对单个资产的敏感属性执行严格解密，并自动还原带有 ENC: 的历史密文
	Decrypt(ctx context.Context, res domain.Resource) (domain.Resource, error)

	// DecryptMany 批量对资产执行敏感属性解密
	DecryptMany(ctx context.Context, resources []domain.Resource) ([]domain.Resource, error)

	// DecryptFields 针对明确指定的字段列表执行解密（用于安全属性关闭时的存量数据异步清洗与还原）
	DecryptFields(ctx context.Context, resources []domain.Resource, fields []string) ([]domain.Resource, error)

	// Mask 将单个资产的敏感字段替换为 "[已脱敏]"，用于展示与审计
	Mask(ctx context.Context, res domain.Resource) (domain.Resource, error)

	// DecryptValue 对单个密文字符串进行解密
	DecryptValue(encryptedText string) (string, error)

	// IsMasked 判断值是否为脱敏占位文本
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

func (p *resourceProtector) DecryptValue(encryptedText string) (string, error) {
	return p.decryptString(encryptedText, "", "")
}

// Encrypt 对单个资产进行敏感属性加密
func (p *resourceProtector) Encrypt(ctx context.Context, res domain.Resource) (domain.Resource, error) {
	if len(res.Data) == 0 {
		return res, nil
	}

	secureFields, err := p.getSecureFields(ctx, res.ModelUID)
	if err != nil || len(secureFields) == 0 {
		return res, err
	}

	res.Data = p.encryptData(res.Data, secureFields, res.ModelUID)
	return res, nil
}

// EncryptMany 批量对资产进行敏感属性加密（聚合查询各模型敏感字段，避免 N 次 RPC）
func (p *resourceProtector) EncryptMany(ctx context.Context, resources []domain.Resource) ([]domain.Resource, error) {
	if len(resources) == 0 {
		return resources, nil
	}

	secureFieldsMap, err := p.getSecureFieldsMap(ctx, resources)
	if err != nil {
		return nil, err
	}

	return lo.Map(resources, func(res domain.Resource, _ int) domain.Resource {
		secureFields := secureFieldsMap[res.ModelUID]
		if len(secureFields) == 0 || len(res.Data) == 0 {
			return res
		}
		res.Data = p.encryptData(res.Data, secureFields, res.ModelUID)
		return res
	}), nil
}

// Decrypt 对单个资产进行敏感属性解密（含 ENC: 兜底解密）
func (p *resourceProtector) Decrypt(ctx context.Context, res domain.Resource) (domain.Resource, error) {
	if len(res.Data) == 0 {
		return res, nil
	}

	secureFields, err := p.getSecureFields(ctx, res.ModelUID)
	if err != nil {
		return res, err
	}

	res.Data = p.decryptData(res.Data, secureFields, res.ModelUID)
	return res, nil
}

// DecryptMany 批量对资产进行敏感属性解密
func (p *resourceProtector) DecryptMany(ctx context.Context, resources []domain.Resource) ([]domain.Resource, error) {
	if len(resources) == 0 {
		return resources, nil
	}

	secureFieldsMap, err := p.getSecureFieldsMap(ctx, resources)
	if err != nil {
		return nil, err
	}

	return lo.Map(resources, func(res domain.Resource, _ int) domain.Resource {
		if len(res.Data) == 0 {
			return res
		}
		res.Data = p.decryptData(res.Data, secureFieldsMap[res.ModelUID], res.ModelUID)
		return res
	}), nil
}

// DecryptFields 对资源切片中指定的字段执行解密
func (p *resourceProtector) DecryptFields(ctx context.Context, resources []domain.Resource, fields []string) ([]domain.Resource, error) {
	if len(resources) == 0 || len(fields) == 0 {
		return resources, nil
	}

	return lo.Map(resources, func(res domain.Resource, _ int) domain.Resource {
		if len(res.Data) == 0 {
			return res
		}
		res.Data = maps.Clone(res.Data)
		for _, field := range fields {
			if strVal, ok := res.Data[field].(string); ok && strVal != "" {
				if decrypted, err := p.decryptString(strVal, res.ModelUID, field); err == nil {
					res.Data[field] = decrypted
				} else {
					p.logger.Warn("指定字段解密失败，保留原值", elog.String("field", field), elog.FieldErr(err))
				}
			}
		}
		return res
	}), nil
}

// Mask 将资产中的敏感属性值替换为统一展示掩码
func (p *resourceProtector) Mask(ctx context.Context, res domain.Resource) (domain.Resource, error) {
	if len(res.Data) == 0 {
		return res, nil
	}

	secureFields, err := p.getSecureFields(ctx, res.ModelUID)
	if err != nil || len(secureFields) == 0 {
		return res, err
	}

	res.Data = maps.Clone(res.Data)
	for _, field := range secureFields {
		if strVal, ok := res.Data[field].(string); ok && strVal != "" {
			res.Data[field] = cryptox.DefaultMask
		}
	}

	return res, nil
}

// 内部加解密数据辅助方法

func (p *resourceProtector) encryptData(data map[string]any, secureFields []string, modelUID string) map[string]any {
	result := maps.Clone(data)

	for _, field := range secureFields {
		if strVal, ok := result[field].(string); ok && strVal != "" && !p.IsMasked(strVal) {
			if encrypted, err := p.protector.Encrypt(strVal); err == nil {
				result[field] = encrypted
			} else {
				p.logger.Error("资产属性加密失败", elog.String("model_uid", modelUID), elog.String("field", field), elog.FieldErr(err))
			}
		}
	}

	return result
}

func (p *resourceProtector) decryptData(data map[string]any, secureFields []string, modelUID string) map[string]any {
	result := maps.Clone(data)

	// 1. 解密已配置为敏感的安全字段
	for _, field := range secureFields {
		if strVal, ok := result[field].(string); ok && strVal != "" {
			if decrypted, err := p.decryptString(strVal, modelUID, field); err == nil {
				result[field] = decrypted
			}
		}
	}

	// 2. 兜底解密：即使当前字段非安全属性，若内容为 ENC: 格式的历史密文，自动还原为明文展示
	for field, val := range result {
		if lo.Contains(secureFields, field) {
			continue
		}
		if strVal, ok := val.(string); ok && strings.HasPrefix(strVal, cryptox.EncryptedPrefix) {
			if decrypted, err := p.decryptString(strVal, modelUID, field); err == nil {
				result[field] = decrypted
			}
		}
	}

	return result
}

func (p *resourceProtector) decryptString(strVal string, modelUID, field string) (string, error) {
	decrypted, err := p.protector.DecryptCiphertext(strVal)
	if err != nil {
		decrypted, err = p.protector.Decrypt(strVal)
		if err != nil {
			p.logger.Error("资产属性解密失败", elog.String("model_uid", modelUID), elog.String("field", field), elog.FieldErr(err))
			return "", fmt.Errorf("资产字段 %s 解密失败: %w", field, err)
		}
	}
	return decrypted, nil
}

func (p *resourceProtector) getSecureFields(ctx context.Context, modelUID string) ([]string, error) {
	if p.attrSvc == nil || modelUID == "" {
		return nil, nil
	}
	secureFieldsMap, err := p.attrSvc.SearchAttributeFieldsBySecure(ctx, []string{modelUID})
	if err != nil {
		return nil, fmt.Errorf("查询模型 %s 敏感属性失败: %w", modelUID, err)
	}
	return secureFieldsMap[modelUID], nil
}

func (p *resourceProtector) getSecureFieldsMap(ctx context.Context, resources []domain.Resource) (map[string][]string, error) {
	if p.attrSvc == nil {
		return map[string][]string{}, nil
	}

	modelUIDs := lo.Uniq(lo.FilterMap(resources, func(r domain.Resource, _ int) (string, bool) {
		return r.ModelUID, r.ModelUID != ""
	}))
	if len(modelUIDs) == 0 {
		return map[string][]string{}, nil
	}

	return p.attrSvc.SearchAttributeFieldsBySecure(ctx, modelUIDs)
}
