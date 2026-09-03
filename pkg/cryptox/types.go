package cryptox

import "errors"

// EncryptedPrefix 是带版本标记的标准密文全局前缀。
const EncryptedPrefix = "ENC:"

// DefaultMask 是敏感值对外展示时使用的统一占位文本。
// 它只表达“该值已被隐藏”，不会泄露原始值的长度或内容。
const DefaultMask = "[已脱敏]"

// ErrInvalidCiphertext 表示输入既不是当前格式密文，也不是已注册的历史密文。
var ErrInvalidCiphertext = errors.New("invalid ciphertext")

// Crypto 定义通用加解密接口。
type Crypto interface {
	// Encrypt 加密明文字符串并返回密文。
	Encrypt(plainText string) (string, error)

	// Decrypt 解密密文字符串并返回明文。
	Decrypt(encryptedText string) (string, error)
}

// Masker 提供敏感值的安全展示形式。
type Masker interface {
	// Mask 返回敏感值的掩码占位形式。
	Mask(value string) string
}

// CiphertextDecryptor 只接受可确认的密文并返回解密后的明文。
// 该接口会兼容没有 ENC:版本: 前缀的历史密文，但遇到非法明文或损坏密文时必须严格阻断。
type CiphertextDecryptor interface {
	DecryptCiphertext(encryptedText string) (string, error)
}

// ValueProtector 统一封装敏感值在存储、运行时和展示边界上的全生命周期处理能力。
type ValueProtector interface {
	Crypto
	CiphertextDecryptor
	Masker
}
