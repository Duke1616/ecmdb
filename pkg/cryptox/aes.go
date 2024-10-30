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

// EncryptAES encrypts any data type using a given key
func EncryptAES[T any](key string, data T) (string, error) {
	plainText, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	paddedKey, err := padKey(key)
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

// DecryptAES decrypts an encrypted string into the original data type
func DecryptAES[T any](key string, encryptedText string) (T, error) {
	var result T
	decodedText, err := hex.DecodeString(encryptedText)
	if err != nil {
		return result, err
	}

	paddedKey, err := padKey(key)
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
