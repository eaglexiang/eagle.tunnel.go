package et

import (
	"net"

	"github.com/eaglexiang/eagle.tunnel.go/src/logger"
	mycipher "github.com/eaglexiang/go-cipher"
	mytunnel "github.com/eaglexiang/go-tunnel"
)

// connect2Relayer 连接到下一个Relayer，完成版本校验和用户校验两个步骤
func (et *ET) connect2Relayer(tunnel *mytunnel.Tunnel) error {
	conn, err := net.DialTimeout("tcp", et.arg.RemoteIPE, et.arg.Timeout)
	if err != nil {
		logger.Warning(err)
		return err
	}
	tunnel.Right = conn
	err = et.checkVersionOfRelayer(tunnel)
	if err != nil {
		return err
	}
	c := mycipher.DefaultCipher()
	if c == nil {
		panic("cipher is nil")
	}
	tunnel.EncryptRight = c.Encrypt
	tunnel.DecryptRight = c.Decrypt
	return et.checkUserOfLocal(tunnel)
}
