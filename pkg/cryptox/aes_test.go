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
