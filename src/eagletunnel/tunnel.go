package eagletunnel

import (
	"net"
	"sync"
	"time"
)

// Tunnel 是一个全双工的双向隧道，内置加密解密、暂停等待的控制器
type Tunnel struct {
	left            *net.Conn
	right           *net.Conn
	encryptLeft     bool
	encryptRight    bool
	bytesL2R        uint
	bytesR2L        uint
	encryptKey      byte
	bufferL2R       chan []byte
	bufferR2L       chan []byte
	l2RIsRunning    bool
	mutexL2R        sync.Mutex
	r2LIsRunning    bool
	mutexR2L        sync.Mutex
	bytesFlowedL2R  int64
	mutexOfBytesL2R sync.Mutex
	bytesFlowedR2L  int64
	mutexOfBytesR2L sync.Mutex
	flowed          bool
	closed          bool
	pause           *bool
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
	tunnel.closed = true
}

// flow from Left 2 Buffer
func (tunnel *Tunnel) flowL2B() {
	var buffer = make([]byte, 1024)
	var count int
	for tunnel.l2RIsRunning {
		if tunnel.pause != nil && *tunnel.pause {
			time.Sleep(time.Second * 1)
		}
		count, _ = tunnel.readLeft(buffer)
		if count > 0 {
			newBuffer := make([]byte, count)
			copy(newBuffer, buffer[:count])
			tunnel.bufferL2R <- newBuffer
			tunnel.mutexOfBytesL2R.Lock()
			tunnel.bytesFlowedL2R += int64(count)
			tunnel.mutexOfBytesL2R.Unlock()
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
		if tunnel.pause != nil && *tunnel.pause {
			time.Sleep(time.Second * 1)
		}
		count, _ = tunnel.readRight(buffer)
		if count > 0 {
			newBuffer := make([]byte, count)
			copy(newBuffer, buffer[:count])
			tunnel.bufferR2L <- newBuffer
			tunnel.mutexOfBytesR2L.Lock()
			tunnel.bytesFlowedR2L += int64(count)
			tunnel.mutexOfBytesR2L.Unlock()
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

	tunnel.flowed = true
}

// BytesFlowed 将返回 bytesFlowedL2R 与 bytesFlowedR2L 的和
func (tunnel *Tunnel) bytesFlowed() int64 {
	var l2r int64
	var r2l int64

	tunnel.mutexOfBytesL2R.Lock()
	l2r = tunnel.bytesFlowedL2R
	tunnel.bytesFlowedL2R = 0
	tunnel.mutexOfBytesL2R.Unlock()

	tunnel.mutexOfBytesR2L.Lock()
	r2l = tunnel.bytesFlowedR2L
	tunnel.bytesFlowedR2L = 0
	tunnel.mutexOfBytesR2L.Unlock()

	return l2r + r2l
}

func (tunnel *Tunnel) isRunning() bool {
	return tunnel.l2RIsRunning || tunnel.r2LIsRunning
}
