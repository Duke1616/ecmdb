package cryptox

import (
	"fmt"
	"strings"
)

type MigrationHandler func(oldEncrypted, newEncrypted string)

// CryptoManager 统一管理多种加密算法，支持多版本并兼容没有版本前缀的历史密文。
type CryptoManager struct {
	algorithms  map[string]Crypto
	defaultVer  string
	legacyVers  []string
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

// WithLegacyAlgo 设置单个历史兼容算法版本（老数据没加 ENC: 前缀时使用）
func (m *CryptoManager) WithLegacyAlgo(version string) *CryptoManager {
	if version != "" {
		for _, registered := range m.legacyVers {
			if registered == version {
				return m
			}
		}
		m.legacyVers = append(m.legacyVers, version)
	}
	return m
}

// WithLegacyAlgos 注册没有版本前缀的历史算法列表，按传入顺序依次尝试解密
func (m *CryptoManager) WithLegacyAlgos(versions ...string) *CryptoManager {
	m.legacyVers = append([]string(nil), versions...)
	return m
}

// WithMigrationHandler 设置迁移回调
func (m *CryptoManager) WithMigrationHandler(handler MigrationHandler) *CryptoManager {
	m.onMigration = handler
	return m
}

// Encrypt 加密数据，自动加上 ENC:<defaultVer>: 前缀
func (m *CryptoManager) Encrypt(plainText string) (string, error) {
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

// Decrypt 解密数据（自动识别版本，兼容历史数据和普通字符串明文）
func (m *CryptoManager) Decrypt(encryptedText string) (string, error) {
	value, err := m.DecryptCiphertext(encryptedText)
	if err != nil {
		// 保留对普通字符串的历史兼容行为；严格校验边界调用 DecryptCiphertext，不能通过这里的宽松行为绕过保护边界。
		if !strings.HasPrefix(encryptedText, EncryptedPrefix) {
			return encryptedText, nil
		}
	}
	return value, err
}

// DecryptCiphertext 根据密文格式选择对应算法进行严格解密。
// 带 ENC:版本: 前缀的数据按指定版本解密；无前缀数据按历史算法列表依次尝试。
// 无法由任何算法解密时严格返回错误，调用方不得将其当作合法密文。
func (m *CryptoManager) DecryptCiphertext(encryptedText string) (string, error) {
	if !strings.HasPrefix(encryptedText, EncryptedPrefix) {
		if val, ok := m.tryLegacyDecrypt(encryptedText); ok {
			return val, nil
		}
		return "", ErrInvalidCiphertext
	}

	trimmed := strings.TrimPrefix(encryptedText, EncryptedPrefix)
	parts := strings.SplitN(trimmed, ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("%w: invalid encrypted format", ErrInvalidCiphertext)
	}

	version, payload := parts[0], parts[1]
	algo, ok := m.algorithms[version]
	if !ok {
		return "", fmt.Errorf("unsupported encryption version: %s", version)
	}

	return algo.Decrypt(payload)
}

// 尝试 legacy 解密，成功时触发迁移回调
func (m *CryptoManager) tryLegacyDecrypt(encryptedText string) (string, bool) {
	for _, legacyVer := range m.legacyVers {
		legacy, ok := m.algorithms[legacyVer]
		if legacyVer == "" || !ok {
			continue
		}

		val, err := legacy.Decrypt(encryptedText)
		if err != nil {
			continue
		}

		// 触发自动迁移回调
		if m.onMigration != nil {
			if newEnc, err1 := m.Encrypt(val); err1 == nil {
				m.onMigration(encryptedText, newEnc)
			}
		}
		return val, true
	}
	return "", false
}
