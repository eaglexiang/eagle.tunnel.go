package eagletunnel

import (
	"net"
	"sync"
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
	bufferL2R    chan []byte
	bufferR2L    chan []byte
	l2RIsRunning bool
	mutexL2R     sync.Mutex
	r2LIsRunning bool
	mutexR2L     sync.Mutex
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

func (tunnel *Tunnel) stopL2R() {
	tunnel.mutexL2R.Lock()
	if tunnel.l2RIsRunning {
		tunnel.l2RIsRunning = false
		if tunnel.right != nil {
			(*tunnel.right).Close()
		}
	}
	tunnel.mutexL2R.Unlock()
}

func (tunnel *Tunnel) stopR2L() {
	tunnel.mutexR2L.Lock()
	if tunnel.r2LIsRunning {
		tunnel.r2LIsRunning = false
		if tunnel.left != nil {
			(*tunnel.left).Close()
		}
	}
	tunnel.mutexR2L.Unlock()
}

func (tunnel *Tunnel) close() {
	tunnel.stopL2R()
	tunnel.stopR2L()
}

// flow from Left 2 Buffer
func (tunnel *Tunnel) flowL2B() {
	var buffer = make([]byte, 1024)
	var count int
	for tunnel.l2RIsRunning {
		count, _ = tunnel.readLeft(buffer)
		if count > 0 {
			newBuffer := make([]byte, count)
			copy(newBuffer, buffer[:count])
			tunnel.bufferL2R <- newBuffer
		} else {
			break
		}
	}
	close(tunnel.bufferL2R)
}

// flow from buffer 2 Right
func (tunnel *Tunnel) flowB2R() {
	var buffer []byte
	var count int
	var ok bool
	for {
		buffer, ok = <-tunnel.bufferL2R
		if ok {
			count, _ = tunnel.writeRight(buffer)
			if count <= 0 {
				break
			}
		} else {
			break
		}
	}
	tunnel.stopL2R()
}

func (tunnel *Tunnel) flowL2R() {
	tunnel.bufferL2R = make(chan []byte, 1024)
	go tunnel.flowL2B()
	go tunnel.flowB2R()
}

func (tunnel *Tunnel) flowR2B() {
	var buffer = make([]byte, 1024)
	var count int
	for tunnel.r2LIsRunning {
		count, _ = tunnel.readRight(buffer)
		if count > 0 {
			newBuffer := make([]byte, count)
			copy(newBuffer, buffer[:count])
			tunnel.bufferR2L <- newBuffer
		} else {
			break
		}
	}
	close(tunnel.bufferR2L)
}

func (tunnel *Tunnel) flowB2L() {
	var buffer []byte
	var count int
	var ok bool
	for {
		buffer, ok = <-tunnel.bufferR2L
		if ok {
			count, _ = tunnel.writeLeft(buffer)
			if count <= 0 {
				break
			}
		} else {
			break
		}
	}
	tunnel.stopR2L()
}

func (tunnel *Tunnel) flowR2L() {
	tunnel.bufferR2L = make(chan []byte, 1024)
	go tunnel.flowR2B()
	go tunnel.flowB2L()
}

func (tunnel *Tunnel) flow() {
	tunnel.mutexL2R.Lock()
	if !tunnel.l2RIsRunning {
		tunnel.l2RIsRunning = true
		go tunnel.flowL2R()
	}
	tunnel.mutexL2R.Unlock()

	tunnel.mutexR2L.Lock()
	if !tunnel.r2LIsRunning {
		tunnel.r2LIsRunning = true
		go tunnel.flowR2L()
	}
	tunnel.mutexR2L.Unlock()
}
