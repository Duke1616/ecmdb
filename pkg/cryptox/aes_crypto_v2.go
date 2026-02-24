package cryptox

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
)

// CryptoAESV2 AES V2 加密器实现
type CryptoAESV2 struct {
	aead cipher.AEAD
}

// NewAESCryptoV2 创建安全强化的 V2 AES 加密器 (不带 JSON，且固定 SHA256 为 32 字节密钥)
func NewAESCryptoV2(key string) (*CryptoAESV2, error) {
	hash := sha256.Sum256([]byte(key))

	block, err := aes.NewCipher(hash[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &CryptoAESV2{
		aead: gcm,
	}, nil
}

// MustNewAESCryptoV2 包装 NewAESCryptoV2，出错时直接 panic (配合 Builder 模式使用)
func MustNewAESCryptoV2(key string) Crypto {
	algo, err := NewAESCryptoV2(key)
	if err != nil {
		panic("failed to init v2 aes crypto: " + err.Error())
	}
	return algo
}

// Encrypt 加密
func (a *CryptoAESV2) Encrypt(plainText string) (string, error) {
	nonce := make([]byte, a.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// V2 干净的字节
	ptBytes := []byte(plainText)

	// aead.Seal 是并发安全的
	ciphertext := a.aead.Seal(nonce, nonce, ptBytes, nil)
	return hex.EncodeToString(ciphertext), nil
}

// Decrypt 解密
func (a *CryptoAESV2) Decrypt(encryptedText string) (string, error) {
	decodedText, err := hex.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	nonceSize := a.aead.NonceSize()
	if len(decodedText) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := decodedText[:nonceSize], decodedText[nonceSize:]
	ptBytes, err := a.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	// V2 干净还原
	return string(ptBytes), nil
}
