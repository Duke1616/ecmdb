package cryptox

// EncryptedPrefix 全局前缀
const EncryptedPrefix = "ENC:"

// Crypto 定义加密解密接口
type Crypto[T any] interface {
	// Encrypt 加密任意类型的数据
	Encrypt(data T) (string, error)

	// Decrypt 解密数据到指定类型
	Decrypt(encryptedText string) (T, error)
}
