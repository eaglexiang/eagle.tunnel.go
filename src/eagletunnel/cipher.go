package eagletunnel

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
)

// Cipher 加解密器
type Cipher struct {
	key []byte
	iv  []byte
	en  *cipher.Stream
	de  *cipher.Stream
}

// SetPassword 设置密码
func (c *Cipher) SetPassword(password string) {
	key := md5.Sum([]byte(password))
	c.key = key[:]
	if c.iv == nil {
		c.iv = c.key
	}
	c.reset()
}

// SetIV 设置IV
func (c *Cipher) SetIV(iv []byte) {
	c.iv = iv
	c.reset()
}

func (c *Cipher) reset() {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		panic(err)
	}
	en := cipher.NewCFBEncrypter(block, c.iv)
	block, err = aes.NewCipher(c.key)
	if err != nil {
		panic(err)
	}
	de := cipher.NewCFBDecrypter(block, c.iv)
	c.en = &en
	c.de = &de
}

// Encrypt 加密
func (c *Cipher) Encrypt(data []byte) {
	(*c.en).XORKeyStream(data, data)
}

// Decrypt 解密
func (c *Cipher) Decrypt(data []byte) {
	(*c.de).XORKeyStream(data, data)
}

// // Cipher 加解密器
// type Cipher struct {
// 	key   []byte
// 	enInd int
// 	deInd int
// }

// // CreateCipher 根据密钥构建Cipher
// func CreateCipher(password string) *Cipher {
// 	md5Sum := md5.Sum([]byte(password))
// 	key := []byte(md5Sum[:])
// 	return &Cipher{key: key}
// }

// // Encrypt 加密
// func (c *Cipher) Encrypt(data []byte) {
// 	for i := 0; i < len(data); i++ {
// 		data[i] = data[i] ^ c.key[c.enInd%len(c.key)]
// 		c.enInd++
// 	}
// }

// // Decrypt 解密
// func (c *Cipher) Decrypt(data []byte) {
// 	for i := 0; i < len(data); i++ {
// 		data[i] = data[i] ^ c.key[c.deInd%len(c.key)]
// 		c.deInd++
// 	}
// }
