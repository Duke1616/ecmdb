package cryptox

import (
	"fmt"
	"testing"
)

func TestDecryptAES(t *testing.T) {
	key := "1234567890" // Key must be 16, 24, or 32 bytes long
	data := map[string]interface{}{
		"username": "user1",
		"password": "pass1",
	}

	// Encrypt
	encrypted, err := EncryptAES(key, data)
	if err != nil {
		fmt.Println("Encryption error:", err)
		return
	}
	fmt.Println("Encrypted:", encrypted)

	// Decrypt
	decryptedData, err := DecryptAES[any](key, encrypted)
	if err != nil {
		fmt.Println("Decryption error:", err)
		return
	}
	fmt.Println("Decrypted:", decryptedData)
}

func TestAESCryptoInterface(t *testing.T) {
	key := "1234567890"
	data := map[string]interface{}{
		"username": "user1",
		"password": "pass1",
	}

	// 使用接口
	crypto := NewAESCrypto[map[string]interface{}](key)

	// Encrypt
	encrypted, err := crypto.Encrypt(data)
	if err != nil {
		t.Fatalf("Encryption error: %v", err)
	}
	fmt.Println("Interface Encrypted:", encrypted)

	// Decrypt
	decryptedData, err := crypto.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decryption error: %v", err)
	}
	fmt.Println("Interface Decrypted:", decryptedData)
}

func TestAESCryptoString(t *testing.T) {
	key := "1234567890"
	data := "hello world"

	// 使用接口
	crypto := NewAESCrypto[string](key)

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
