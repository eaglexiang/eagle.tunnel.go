package eagletunnel

import (
	"net"
)

const (
	Left = iota
	Right
)

type Tunnel struct {
	left         *net.Conn
	right        *net.Conn
	encryptLeft  bool
	encryptRight bool
	bytesL2R     uint
	bytesR2L     uint
	encryptKey   byte
}

func (tunnel *Tunnel) encrypt(data []byte) {
	for i, value := range data {
		data[i] = value ^ tunnel.encryptKey
	}
}

func (tunnel *Tunnel) decrypt(data []byte) {
	for i, value := range data {
		data[i] = value ^ tunnel.encryptKey
	}
}

func (tunnel *Tunnel) writeLeft(data []byte) (int, error) {
	if tunnel.encryptLeft {
		tunnel.encrypt(data)
	}
	count, err := (*tunnel.left).Write(data)
	return count, err
}

func (tunnel *Tunnel) writeRight(data []byte) (int, error) {
	if tunnel.encryptRight {
		tunnel.encrypt(data)
	}
	count, err := (*tunnel.right).Write(data)
	return count, err
}

func (tunnel *Tunnel) readLeft(data []byte) (int, error) {
	count, err := (*tunnel.left).Read(data)
	if err == nil {
		if tunnel.encryptLeft {
			tunnel.decrypt(data)
		}
	}
	return count, err
}

func (tunnel *Tunnel) readRight(data []byte) (int, error) {
	count, err := (*tunnel.right).Read(data)
	if err == nil {
		if tunnel.encryptRight {
			tunnel.decrypt(data)
		}
	}
	return count, err
}

func (tunnel *Tunnel) close() {
	if tunnel.left != nil {
		(*tunnel.left).Close()
	}
	if tunnel.right != nil {
		(*tunnel.right).Close()
	}
}

func (tunnel *Tunnel) flowL2R() {
	var buffer = make([]byte, 1024)
	var count int
	for {
		count, _ = tunnel.readLeft(buffer)
		if count > 0 {
			count, _ = tunnel.writeRight(buffer[0:count])
			if count > 0 {
				continue
			} else {
				tunnel.close()
				break
			}
		} else {
			tunnel.close()
			break
		}
	}
}

func (tunnel *Tunnel) flowR2L() {
	var buffer = make([]byte, 1024)
	var count int
	for {
		count, _ = tunnel.readRight(buffer)
		if count > 0 {
			count, _ = tunnel.writeLeft(buffer[0:count])
			if count > 0 {
				continue
			} else {
				tunnel.close()
				break
			}
		} else {
			tunnel.close()
			break
		}
	}
}

func (tunnel *Tunnel) flow() {
	go tunnel.flowL2R()
	tunnel.flowR2L()
}
