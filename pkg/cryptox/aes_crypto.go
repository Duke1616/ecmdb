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

// CryptoAES AES 加密器实现
type CryptoAES[T any] struct {
	key string
}

// NewAESCrypto 创建新的 AES 加密器
func NewAESCrypto[T any](key string) Crypto[T] {
	return &CryptoAES[T]{key: key}
}

// Encrypt 加密任意类型的数据
func (a *CryptoAES[T]) Encrypt(data T) (string, error) {
	plainText, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	paddedKey, err := padKey(a.key)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(paddedKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, plainText, nil)
	return hex.EncodeToString(ciphertext), nil
}

// Decrypt 解密数据到指定类型
func (a *CryptoAES[T]) Decrypt(encryptedText string) (T, error) {
	var result T
	decodedText, err := hex.DecodeString(encryptedText)
	if err != nil {
		return result, err
	}

	paddedKey, err := padKey(a.key)
	if err != nil {
		return result, err
	}

	block, err := aes.NewCipher(paddedKey)
	if err != nil {
		return result, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return result, err
	}

	nonceSize := gcm.NonceSize()
	if len(decodedText) < nonceSize {
		return result, errors.New("ciphertext too short")
	}

	nonce, ciphertext := decodedText[:nonceSize], decodedText[nonceSize:]
	plainText, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(plainText, &result)
	return result, err
}

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
