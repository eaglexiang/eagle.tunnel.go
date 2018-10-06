package eagletunnel

import (
	"crypto/md5"
)

// // Cipher 加解密器
// type Cipher struct {
// 	en *cipher.Stream
// 	de *cipher.Stream
// }

// // CreateCipher 根据密码创建新Cipher
// func CreateCipher(password string) *Cipher {
// 	key := md5.Sum([]byte(password))
// 	iv := key[:]
// 	block, err := aes.NewCipher([]byte(key[:]))
// 	if err != nil {
// 		panic(err)
// 	}
// 	en := cipher.NewCFBEncrypter(block, []byte(iv))
// 	block, err = aes.NewCipher([]byte(key[:]))
// 	if err != nil {
// 		panic(err)
// 	}
// 	de := cipher.NewCFBDecrypter(block, []byte(iv))
// 	return &Cipher{en: &en, de: &de}
// }

// // Encrypt 加密
// func (c *Cipher) Encrypt(data []byte) {
// 	(*c.en).XORKeyStream(data, data)
// }

// // Decrypt 解密
// func (c *Cipher) Decrypt(data []byte) {
// 	(*c.de).XORKeyStream(data, data)
// }

// Cipher 加解密器
type Cipher struct {
	key   []byte
	enInd int
	deInd int
}

// CreateCipher 根据密钥构建Cipher
func CreateCipher(password string) *Cipher {
	md5Sum := md5.Sum([]byte(password))
	key := []byte(md5Sum[:])
	return &Cipher{key: key}
}

// Encrypt 加密
func (c *Cipher) Encrypt(data []byte) {
	for i := 0; i < len(data); i++ {
		data[i] = data[i] ^ c.key[c.enInd%len(c.key)]
		c.enInd++
	}
}

// Decrypt 解密
func (c *Cipher) Decrypt(data []byte) {
	for i := 0; i < len(data); i++ {
		data[i] = data[i] ^ c.key[c.deInd%len(c.key)]
		c.deInd++
	}
}
