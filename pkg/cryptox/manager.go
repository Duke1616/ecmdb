package cryptox

import (
	"fmt"
	"strings"
)

type MigrationHandler func(oldEnc, newEnc string)

// CryptoManager 统一管理多种加密算法（支持多版本）
type CryptoManager struct {
	algorithms  map[string]Crypto // 支持多版本
	defaultVer  string            // 默认加密版本
	legacyVer   string            // 历史算法版本（兼容旧 CryptoAES）
	onMigration MigrationHandler
}

// NewCryptoManager 创建新的 CryptoManager
func NewCryptoManager(defaultVer string) *CryptoManager {
	return &CryptoManager{
		algorithms: make(map[string]Crypto),
		defaultVer: defaultVer,
	}
}

// Register 注册符合 Crypto 接口的加密算法策略，绑定到一个版本号
func (m *CryptoManager) Register(version string, algo Crypto) *CryptoManager {
	m.algorithms[version] = algo
	return m
}

// WithLegacyAlgo 设置历史兼容算法（老数据没加 ENC: 前缀时使用）
func (m *CryptoManager) WithLegacyAlgo(version string) *CryptoManager {
	m.legacyVer = version
	return m
}

// WithMigrationHandler 设置迁移回调
func (m *CryptoManager) WithMigrationHandler(handler MigrationHandler) *CryptoManager {
	m.onMigration = handler
	return m
}

// Encrypt 加密
// 自动给旧数据加上 ENC:<version>: 前缀
func (m *CryptoManager) Encrypt(plainText string) (string, error) {
	// HACK: 这里不能做防呆机制，避免明文正好带有前缀导致未加密。
	algo, ok := m.algorithms[m.defaultVer]
	if !ok {
		return "", fmt.Errorf("no default algorithm registered")
	}

	encrypted, err := algo.Encrypt(plainText)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%s:%s", EncryptedPrefix, m.defaultVer, encrypted), nil
}

// Decrypt 解密（自动识别版本，兼容历史数据）
func (m *CryptoManager) Decrypt(encryptedText string) (string, error) {
	// case 1: 无前缀 => 可能是 legacy 或明文
	if !strings.HasPrefix(encryptedText, EncryptedPrefix) {
		// 尝试 legacy 解密
		if val, ok := m.tryLegacyDecrypt(encryptedText); ok {
			return val, nil
		}

		// 尝试当作明文
		return encryptedText, nil
	}

	// case 2: 有前缀 => 新格式
	return m.decryptWithVersion(encryptedText)
}

// 尝试 legacy 解密，成功时触发迁移
func (m *CryptoManager) tryLegacyDecrypt(encryptedText string) (string, bool) {
	if m.legacyVer == "" {
		return "", false
	}
	legacy, ok := m.algorithms[m.legacyVer]
	if !ok {
		return "", false
	}

	val, err := legacy.Decrypt(encryptedText)
	if err != nil {
		return "", false
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
func (m *CryptoManager) decryptWithVersion(encryptedText string) (string, error) {
	trimmed := strings.TrimPrefix(encryptedText, EncryptedPrefix)
	parts := strings.SplitN(trimmed, ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid encrypted format")
	}

	version, payload := parts[0], parts[1]
	algo, ok := m.algorithms[version]
	if !ok {
		return "", fmt.Errorf("unsupported encryption version: %s", version)
	}

	return algo.Decrypt(payload)
}
