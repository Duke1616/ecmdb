package cryptox

import (
	"fmt"
	"testing"
)

func TestDecryptAES(t *testing.T) {
	key := "1234567890" // Key must be at least 1 byte, we use SHA256 internally now!
	data := "my-secret-data"

	// Encrypt
	encrypted, err := EncryptAES(key, data)
	if err != nil {
		fmt.Println("Encryption error:", err)
		return
	}
	fmt.Println("Encrypted:", encrypted)

	// Decrypt
	decryptedData, err := DecryptAES(key, encrypted)
	if err != nil {
		fmt.Println("Decryption error:", err)
		return
	}
	fmt.Println("Decrypted:", decryptedData)
}

func TestAESCryptoString(t *testing.T) {
	key := "1234567890"
	data := "hello world"

	// 使用接口
	crypto, err := NewAESCrypto(key)
	if err != nil {
		t.Fatalf("Init error: %v", err)
	}

	// Encrypt
	encrypted, err := crypto.Encrypt(data)
	if err != nil {
		t.Fatalf("Encryption error: %v", err)
	}
	fmt.Println("String Encrypted:", encrypted)

	// Decrypt
	decryptedData, err := crypto.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decryption error: %v", err)
	}
	fmt.Println("String Decrypted:", decryptedData)
}
