package eagletunnel

import (
	"errors"
	"strconv"
)

// SimpleCipher 传统的简单加密
type SimpleCipher struct {
	dataKey byte
}

// SetPassword 设置密码
func (sc *SimpleCipher) SetPassword(key string) error {
	if key == "" {
		return errors.New("key is empty")
	}
	_dataKey, err := strconv.ParseInt(key, 10, 8)
	if err != nil {
		return err
	}
	sc.dataKey = byte(_dataKey)
	return nil
}

// Encrypt 加密
func (sc *SimpleCipher) Encrypt(data []byte) {
	for i, value := range data {
		data[i] = value ^ sc.dataKey
	}
}

// Decrypt 解密
func (sc *SimpleCipher) Decrypt(data []byte) {
	for i, value := range data {
		data[i] = value ^ sc.dataKey
	}
}
