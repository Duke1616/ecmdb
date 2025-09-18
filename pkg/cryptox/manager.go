package cryptox

import (
	"fmt"
	"strings"
)

type MigrationHandler func(oldEnc, newEnc string)

// CryptoManager 统一管理多种加密算法（支持多版本）
type CryptoManager[T any] struct {
	algorithms  map[string]Crypto[T] // 支持多版本
	defaultVer  string               // 默认加密版本
	legacyVer   string               // 历史算法版本（兼容旧 CryptoAES）
	onMigration MigrationHandler
}

// NewCryptoManager 创建新的 CryptoManager
func NewCryptoManager[T any](defaultVer string) *CryptoManager[T] {
	return &CryptoManager[T]{
		algorithms: make(map[string]Crypto[T]),
		defaultVer: defaultVer,
	}
}

// RegisterAesAlgorithm 注册 AES 算法，带版本号
func (m *CryptoManager[T]) RegisterAesAlgorithm(version, key string) *CryptoManager[T] {
	m.algorithms[version] = NewAESCrypto[T](key)
	return m
}

// WithLegacyAlgo 设置历史兼容算法（老数据没加 ENC: 前缀时使用）
func (m *CryptoManager[T]) WithLegacyAlgo(version string) *CryptoManager[T] {
	m.legacyVer = version
	return m
}

// WithMigrationHandler 设置迁移回调
func (m *CryptoManager[T]) WithMigrationHandler(handler MigrationHandler) *CryptoManager[T] {
	m.onMigration = handler
	return m
}

// Encrypt 加密（幂等，避免重复加密）
// 自动给旧数据加上 ENC:<version>: 前缀
func (m *CryptoManager[T]) Encrypt(data T) (string, error) {
	// 幂等处理，如果 data 已经是 ENC:<version>:<payload>，直接返回
	if s, ok := any(data).(string); ok && strings.HasPrefix(s, EncryptedPrefix) {
		return s, nil
	}

	algo, ok := m.algorithms[m.defaultVer]
	if !ok {
		return "", fmt.Errorf("no default algorithm registered")
	}

	encrypted, err := algo.Encrypt(data)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%s:%s", EncryptedPrefix, m.defaultVer, encrypted), nil
}

// Decrypt 解密（自动识别版本，兼容历史数据）
func (m *CryptoManager[T]) Decrypt(encryptedText string) (T, error) {
	var zero T

	// case 1: 无前缀 => 可能是 legacy 或明文
	if !strings.HasPrefix(encryptedText, EncryptedPrefix) {
		// 尝试 legacy 解密
		if val, ok := m.tryLegacyDecrypt(encryptedText); ok {
			return val, nil
		}

		// 尝试当作明文
		if v, ok := any(encryptedText).(T); ok {
			return v, nil
		}

		return zero, fmt.Errorf("unsupported legacy/plain format: %v", encryptedText)
	}

	// case 2: 有前缀 => 新格式
	return m.decryptWithVersion(encryptedText)
}

// 尝试 legacy 解密，成功时触发迁移
func (m *CryptoManager[T]) tryLegacyDecrypt(encryptedText string) (T, bool) {
	var zero T

	if m.legacyVer == "" {
		return zero, false
	}
	legacy, ok := m.algorithms[m.legacyVer]
	if !ok {
		return zero, false
	}

	val, err := legacy.Decrypt(encryptedText)
	if err != nil {
		return zero, false
	}

	// ⚡ 自动迁移
	if m.onMigration != nil {
		if newEnc, err1 := m.Encrypt(val); err1 == nil {
			m.onMigration(encryptedText, newEnc)
		}
	}
	return val, true
}

// 解密带版本前缀的数据
func (m *CryptoManager[T]) decryptWithVersion(encryptedText string) (T, error) {
	var zero T

	trimmed := strings.TrimPrefix(encryptedText, EncryptedPrefix)
	parts := strings.SplitN(trimmed, ":", 2)
	if len(parts) != 2 {
		return zero, fmt.Errorf("invalid encrypted format")
	}

	version, payload := parts[0], parts[1]
	algo, ok := m.algorithms[version]
	if !ok {
		return zero, fmt.Errorf("unsupported encryption version: %s", version)
	}

	return algo.Decrypt(payload)
}
