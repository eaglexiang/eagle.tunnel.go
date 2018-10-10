package eagletunnel

const (
	// SimpleCipherType 简单加密
	SimpleCipherType = iota
	// AES128CipherType AES128加密
	AES128CipherType
	// UnknownCipherType 未知加密
	UnknownCipherType
)

// Cipher 加密器
type Cipher interface {
	Encrypt([]byte)
	Decrypt([]byte)
	SetPassword(string) error
}

// ParseCipherType 将字符串转化为加密类型
func ParseCipherType(txt string) int {
	switch txt {
	case "simple", "SIMPLE":
		return SimpleCipherType
	case "aes128", "AES128":
		return AES128CipherType
	default:
		return UnknownCipherType
	}
}
