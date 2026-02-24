package cryptox

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
)

// CryptoAES AES 加密器实现 (遗留 V1)
type CryptoAES struct {
	aead cipher.AEAD
}

// padKey 兼容之前的密钥填充逻辑
func padKey(key string) ([]byte, error) {
	keyLength := 16
	if len(key) > keyLength {
		return []byte(key[:keyLength]), nil
	}

	if len(key) < keyLength {
		paddedKey := make([]byte, keyLength)
		copy(paddedKey, key)
		return paddedKey, nil
	}

	return []byte(key), nil
}

// NewAESCrypto 创建 AES 加密器 (遗留 V1 版本，带 JSON 序列化和截断密钥)
func NewAESCrypto(key string) (*CryptoAES, error) {
	paddedKey, err := padKey(key)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(paddedKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &CryptoAES{
		aead: gcm,
	}, nil
}

// MustNewAESCrypto 包装 NewAESCrypto，出错时直接 panic (配合 Builder 模式使用)
func MustNewAESCrypto(key string) Crypto {
	algo, err := NewAESCrypto(key)
	if err != nil {
		panic("failed to init legacy aes crypto: " + err.Error())
	}
	return algo
}

// Encrypt 加密
func (a *CryptoAES) Encrypt(plainText string) (string, error) {
	nonce := make([]byte, a.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// 兼容之前的 json 序列化逻辑，保证产生的密文和旧版本完全一致
	ptBytes, err := json.Marshal(plainText)
	if err != nil {
		return "", err
	}

	// aead.Seal 是并发安全的
	ciphertext := a.aead.Seal(nonce, nonce, ptBytes, nil)
	return hex.EncodeToString(ciphertext), nil
}

// Decrypt 解密
func (a *CryptoAES) Decrypt(encryptedText string) (string, error) {
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

	// 兼容之前的 json 反序列化逻辑
	var result string
	err = json.Unmarshal(ptBytes, &result)
	return result, err
}
