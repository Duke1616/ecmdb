package cryptox

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
)

// CryptoAES 实现与历史密文兼容的 AES-GCM V1 算法（带 JSON 序列化与截断/填充密钥）。
type CryptoAES struct {
	aead cipher.AEAD
}

// NewAESCrypto 创建历史兼容的 V1 AES 加密器。
func NewAESCrypto(key string) (*CryptoAES, error) {
	paddedKey := make([]byte, 16)
	copy(paddedKey, key)
	if len(key) > len(paddedKey) {
		copy(paddedKey, key[:len(paddedKey)])
	}
	block, err := aes.NewCipher(paddedKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &CryptoAES{aead: gcm}, nil
}

// MustNewAESCrypto 包装 NewAESCrypto，初始化失败时直接 panic。
func MustNewAESCrypto(key string) Crypto {
	algorithm, err := NewAESCrypto(key)
	if err != nil {
		panic("failed to init legacy aes crypto: " + err.Error())
	}
	return algorithm
}

func (a *CryptoAES) Encrypt(plainText string) (string, error) {
	plainBytes, err := json.Marshal(plainText)
	if err != nil {
		return "", err
	}
	return seal(a.aead, plainBytes)
}

func (a *CryptoAES) Decrypt(encryptedText string) (string, error) {
	plainBytes, err := open(a.aead, encryptedText)
	if err != nil {
		return "", err
	}
	var result string
	if err = json.Unmarshal(plainBytes, &result); err != nil {
		return "", err
	}
	return result, nil
}

// CryptoAESV2 使用 SHA-256 派生的 32 字节高强度密钥加密原始纯文本。
type CryptoAESV2 struct {
	aead cipher.AEAD
}

// NewAESCryptoV2 创建增强型 V2 AES-GCM 加密器。
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
	return &CryptoAESV2{aead: gcm}, nil
}

// MustNewAESCryptoV2 包装 NewAESCryptoV2，初始化失败时直接 panic。
func MustNewAESCryptoV2(key string) Crypto {
	algorithm, err := NewAESCryptoV2(key)
	if err != nil {
		panic("failed to init v2 aes crypto: " + err.Error())
	}
	return algorithm
}

func (a *CryptoAESV2) Encrypt(plainText string) (string, error) {
	return seal(a.aead, []byte(plainText))
}

func (a *CryptoAESV2) Decrypt(encryptedText string) (string, error) {
	plainBytes, err := open(a.aead, encryptedText)
	if err != nil {
		return "", err
	}
	return string(plainBytes), nil
}

// seal 执行标准 GCM Nonce 随机填充并输出 Hex 密文。
func seal(aead cipher.AEAD, plainText []byte) (string, error) {
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	cipherText := aead.Seal(nonce, nonce, plainText, nil)
	return hex.EncodeToString(cipherText), nil
}

// open 解码 Hex 密文并执行 GCM 校验解密。
func open(aead cipher.AEAD, encryptedText string) ([]byte, error) {
	decoded, err := hex.DecodeString(encryptedText)
	if err != nil {
		return nil, err
	}
	if len(decoded) < aead.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}
	nonce, cipherText := decoded[:aead.NonceSize()], decoded[aead.NonceSize():]
	return aead.Open(nil, nonce, cipherText, nil)
}
