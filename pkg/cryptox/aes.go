package cryptox

// EncryptAES encrypts any data type using a given key
// 为了向后兼容，保留此函数
func EncryptAES(key string, data string) (string, error) {
	crypto, err := NewAESCrypto(key)
	if err != nil {
		return "", err
	}
	return crypto.Encrypt(data)
}

// DecryptAES decrypts an encrypted string into the original data type
// 为了向后兼容，保留此函数
func DecryptAES(key string, encryptedText string) (string, error) {
	crypto, err := NewAESCrypto(key)
	if err != nil {
		return "", err
	}
	return crypto.Decrypt(encryptedText)
}
