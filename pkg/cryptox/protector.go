package cryptox

import "strings"

type valueProtector struct {
	Crypto
	mask string
}

// NewValueProtector 为已有加密组件增加统一的展示掩码与严格解密能力。
func NewValueProtector(cipher Crypto) ValueProtector {
	if cipher == nil {
		return nil
	}
	if protector, ok := cipher.(ValueProtector); ok {
		return protector
	}
	return &valueProtector{Crypto: cipher, mask: DefaultMask}
}

func (p *valueProtector) Mask(string) string {
	return p.mask
}

// DecryptCiphertext 将严格密文解密转发给底层加密管理器。
func (p *valueProtector) DecryptCiphertext(encryptedText string) (string, error) {
	decoder, ok := p.Crypto.(CiphertextDecryptor)
	if !ok {
		if !strings.HasPrefix(encryptedText, EncryptedPrefix) {
			return "", ErrInvalidCiphertext
		}
		return p.Crypto.Decrypt(encryptedText)
	}
	return decoder.DecryptCiphertext(encryptedText)
}

// MaskValue 用于不需要解密、只需要生成展示结果的投影边界。
func MaskValue(string) string { return DefaultMask }
