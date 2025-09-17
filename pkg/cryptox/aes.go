package cryptox

// EncryptAES encrypts any data type using a given key
// 为了向后兼容，保留此函数
func EncryptAES[T any](key string, data T) (string, error) {
	crypto := NewAESCrypto[T](key)
	return crypto.Encrypt(data)
}

// DecryptAES decrypts an encrypted string into the original data type
// 为了向后兼容，保留此函数
func DecryptAES[T any](key string, encryptedText string) (T, error) {
	crypto := NewAESCrypto[T](key)
	return crypto.Decrypt(encryptedText)
}
