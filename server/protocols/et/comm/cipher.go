/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-08-25 11:33:14
 * @LastEditTime: 2019-08-25 11:42:27
 */

package comm

import (
	"net"

	"github.com/eaglexiang/go-cipher"
	cipherconn "github.com/eaglexiang/go-cipher-conn"
	"github.com/eaglexiang/go-settings"
)

// NewCipherConn 构造新的CipherConn
func NewCipherConn(base net.Conn) net.Conn {
	e, d := newCipher()
	conn := cipherconn.New(base, d, e)
	return conn
}

func newCipher() (e cipher.StreamEncryptor, d cipher.StreamDecryptor) {
	cType := settings.Get("cipher")
	switch cType {
	case "simple":
		e, d = newSimpleCipher()
	default:
		panic("invalid cipher type: " + cType)
	}
	return
}

func newSimpleCipher() (e cipher.StreamEncryptor, d cipher.StreamDecryptor) {
	k := settings.Get("data-key")

	c := cipher.SimpleCipher{}
	c.SetKey(k)

	e, d = c, c
	return
}
