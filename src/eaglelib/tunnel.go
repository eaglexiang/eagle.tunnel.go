package eaglelib

import (
	"net"
	"sync"
	"time"
)

// Tunnel 是一个全双工的双向隧道，内置加密解密、暂停等待的控制器
type Tunnel struct {
	Left            *net.Conn
	Right           *net.Conn
	EncryptLeft     bool
	EncryptRight    bool
	bytesL2R        uint
	bytesR2L        uint
	EncryptKey      byte
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
	Flowed          bool
	Closed          bool
	Pause           *bool
}

func (tunnel *Tunnel) encrypt(data []byte) {
	for i, value := range data {
		data[i] = value ^ tunnel.EncryptKey
	}
}

func (tunnel *Tunnel) decrypt(data []byte) {
	for i, value := range data {
		data[i] = value ^ tunnel.EncryptKey
	}
}

// WriteLeft 向左边写数据
func (tunnel *Tunnel) WriteLeft(data []byte) (int, error) {
	if tunnel.EncryptLeft {
		tunnel.encrypt(data)
	}
	count, err := (*tunnel.Left).Write(data)
	return count, err
}

// WriteRight 向右边写数据
func (tunnel *Tunnel) WriteRight(data []byte) (int, error) {
	if tunnel.EncryptRight {
		tunnel.encrypt(data)
	}
	count, err := (*tunnel.Right).Write(data)
	return count, err
}

// ReadLeft 从左边读取数据
func (tunnel *Tunnel) ReadLeft(data []byte) (int, error) {
	count, err := (*tunnel.Left).Read(data)
	if err == nil {
		if tunnel.EncryptLeft {
			tunnel.decrypt(data)
		}
	}
	return count, err
}

// ReadRight 从右边读取数据
func (tunnel *Tunnel) ReadRight(data []byte) (int, error) {
	count, err := (*tunnel.Right).Read(data)
	if err == nil {
		if tunnel.EncryptRight {
			tunnel.decrypt(data)
		}
	}
	return count, err
}

func (tunnel *Tunnel) stopL2R() {
	tunnel.mutexL2R.Lock()
	if tunnel.l2RIsRunning {
		tunnel.l2RIsRunning = false
		if tunnel.Right != nil {
			(*tunnel.Right).Close()
		}
	}
	tunnel.mutexL2R.Unlock()
}

func (tunnel *Tunnel) stopR2L() {
	tunnel.mutexR2L.Lock()
	if tunnel.r2LIsRunning {
		tunnel.r2LIsRunning = false
		if tunnel.Left != nil {
			(*tunnel.Left).Close()
		}
	}
	tunnel.mutexR2L.Unlock()
}

// Close 关闭Tunnel，关闭前会停止其双向的流动
func (tunnel *Tunnel) Close() {
	tunnel.stopL2R()
	tunnel.stopR2L()
	tunnel.Closed = true
}

// flow from Left 2 Buffer
func (tunnel *Tunnel) flowL2B() {
	var buffer = make([]byte, 1024)
	var count int
	for tunnel.l2RIsRunning {
		if tunnel.Pause != nil && *tunnel.Pause {
			time.Sleep(time.Second * 1)
		}
		count, _ = tunnel.ReadLeft(buffer)
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
			count, _ = tunnel.WriteRight(buffer)
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
		if tunnel.Pause != nil && *tunnel.Pause {
			time.Sleep(time.Second * 1)
		}
		count, _ = tunnel.ReadRight(buffer)
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
			count, _ = tunnel.WriteLeft(buffer)
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

// Flow 开始双向流动
func (tunnel *Tunnel) Flow() {
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

	tunnel.Flowed = true
}

// BytesFlowed 将返回 bytesFlowedL2R 与 bytesFlowedR2L 的和
func (tunnel *Tunnel) BytesFlowed() int64 {
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

// IsRunning 至少有一个方向正在流动
func (tunnel *Tunnel) IsRunning() bool {
	return tunnel.l2RIsRunning || tunnel.r2LIsRunning
}
