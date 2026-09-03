package cryptox

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAESCrypto_V1(t *testing.T) {
	key := "my-secret-key-123"
	data := "hello-cmdb-v1"

	crypto := MustNewAESCrypto(key)

	encrypted, err := crypto.Encrypt(data)
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := crypto.Decrypt(encrypted)
	assert.NoError(t, err)
	assert.Equal(t, data, decrypted)

	// 错误密文解密
	_, err = crypto.Decrypt("invalid_hex")
	assert.Error(t, err)

	_, err = crypto.Decrypt("1234") // 过短密文
	assert.Error(t, err)
}

func TestAESCrypto_V2(t *testing.T) {
	key := "my-high-security-key-456"
	data := "hello-cmdb-v2"

	crypto := MustNewAESCryptoV2(key)

	encrypted, err := crypto.Encrypt(data)
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := crypto.Decrypt(encrypted)
	assert.NoError(t, err)
	assert.Equal(t, data, decrypted)

	// 错误密文解密
	_, err = crypto.Decrypt("invalid_hex")
	assert.Error(t, err)

	_, err = crypto.Decrypt("1234") // 过短密文
	assert.Error(t, err)
}
