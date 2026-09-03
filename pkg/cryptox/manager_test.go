package cryptox

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockCrypto struct {
	encryptFunc func(plainText string) (string, error)
	decryptFunc func(encryptedText string) (string, error)
}

func (m *mockCrypto) Encrypt(plainText string) (string, error) {
	if m.encryptFunc != nil {
		return m.encryptFunc(plainText)
	}
	return "mock_encrypted_" + plainText, nil
}

func (m *mockCrypto) Decrypt(encryptedText string) (string, error) {
	if m.decryptFunc != nil {
		return m.decryptFunc(encryptedText)
	}
	return "mock_decrypted_" + encryptedText, nil
}

func TestCryptoManager_EncryptDecrypt(t *testing.T) {
	testCases := []struct {
		name      string
		plainText string
		expectVer string
	}{
		{
			name:      "正常加密并包含版本前缀",
			plainText: "hello-world",
			expectVer: "V2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mgr := NewCryptoManager("V2").
				Register("V2", &mockCrypto{})

			encrypted, err := mgr.Encrypt(tc.plainText)
			assert.NoError(t, err)
			assert.Equal(t, "ENC:V2:mock_encrypted_hello-world", encrypted)

			decrypted, err := mgr.Decrypt(encrypted)
			assert.NoError(t, err)
			assert.Equal(t, "mock_decrypted_mock_encrypted_hello-world", decrypted)
		})
	}
}

func TestCryptoManager_DecryptCiphertext_FailClosed(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		wantErr     bool
		errExpected error
	}{
		{
			name:        "非法普通明文在严格解密时报错阻断",
			input:       "plain_text_not_encrypted",
			wantErr:     true,
			errExpected: ErrInvalidCiphertext,
		},
		{
			name:        "非法格式前缀密文在严格解密时报错",
			input:       "ENC:invalid_no_colon",
			wantErr:     true,
			errExpected: ErrInvalidCiphertext,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mgr := NewCryptoManager("V2").
				Register("V2", &mockCrypto{})

			val, err := mgr.DecryptCiphertext(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				if tc.errExpected != nil {
					assert.True(t, errors.Is(err, tc.errExpected))
				}
				assert.Empty(t, val)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestCryptoManager_MultiLegacyAlgos(t *testing.T) {
	v1Mock := &mockCrypto{
		decryptFunc: func(encryptedText string) (string, error) {
			if encryptedText == "v1_legacy_cipher" {
				return "legacy_v1_plain", nil
			}
			return "", errors.New("not v1")
		},
	}
	v0Mock := &mockCrypto{
		decryptFunc: func(encryptedText string) (string, error) {
			if encryptedText == "v0_legacy_cipher" {
				return "legacy_v0_plain", nil
			}
			return "", errors.New("not v0")
		},
	}

	migratedOld := ""
	migratedNew := ""

	mgr := NewCryptoManager("V2").
		Register("V2", &mockCrypto{}).
		Register("V1", v1Mock).
		Register("V0", v0Mock).
		WithLegacyAlgos("V1", "V0").
		WithMigrationHandler(func(oldEncrypted, newEncrypted string) {
			migratedOld = oldEncrypted
			migratedNew = newEncrypted
		})

	// 测试 V1 legacy
	val, err := mgr.DecryptCiphertext("v1_legacy_cipher")
	assert.NoError(t, err)
	assert.Equal(t, "legacy_v1_plain", val)
	assert.Equal(t, "v1_legacy_cipher", migratedOld)
	assert.Equal(t, "ENC:V2:mock_encrypted_legacy_v1_plain", migratedNew)

	// 测试 V0 legacy
	val0, err := mgr.DecryptCiphertext("v0_legacy_cipher")
	assert.NoError(t, err)
	assert.Equal(t, "legacy_v0_plain", val0)

	// 测试都不命中的脏数据
	_, errNone := mgr.DecryptCiphertext("dirty_data")
	assert.ErrorIs(t, errNone, ErrInvalidCiphertext)
}

func TestValueProtector_Masking(t *testing.T) {
	mgr := NewCryptoManager("V2").
		Register("V2", &mockCrypto{})

	protector := NewValueProtector(mgr)

	assert.Equal(t, DefaultMask, protector.Mask("sensitive_password"))
	assert.Equal(t, DefaultMask, MaskValue("sensitive_token"))
}
