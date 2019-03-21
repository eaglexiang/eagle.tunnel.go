package et

import (
	"errors"
	"net"
	"strings"

	"github.com/eaglexiang/eagle.tunnel.go/src/logger"
	mycipher "github.com/eaglexiang/go-cipher"
	mytunnel "github.com/eaglexiang/go-tunnel"
)

// ET子协议的类型
const (
	EtUNKNOWN = iota
	EtTCP
	EtDNS
	EtDNS6
	EtLOCATION
	EtCHECK
	EtBIND
)

// ET子协议的名字
const (
	EtNameUNKNOWN  = "UNKNOWN"
	EtNameTCP      = "TCP"
	EtNameDNS      = "DNS"
	EtNameDNS6     = "DNS6"
	EtNameLOCATION = "LOCATION"
	EtNameCHECK    = "CHECK"
	EtNameBIND     = "BIND"
)

// 代理的状态
const (
	ProxyENABLE = iota
	ProxySMART
	ErrorProxyStatus
)

// handler ET子协议的handler接口
type handler interface {
	Handle(req string, tunnel *mytunnel.Tunnel) error // 处理业务
	Type() int                                        // ET子协议的类型
	Name() string                                     // ET子协议的名字
}

// sender ET子协议的sender
type sender interface {
	Send(et *ET, e *NetArg) error //发送流程
	Type() int                    // ET子协议的类型
	Name() string                 // ET子协议的名字
}

// EtTypes ET子协议的类型
var EtTypes map[string]int

// EtNames ET子协议的名字
var EtNames map[int]string

func init() {
	EtTypes = make(map[string]int)
	EtTypes[EtNameTCP] = EtTCP
	EtTypes[EtNameDNS] = EtDNS
	EtTypes[EtNameDNS6] = EtDNS6
	EtTypes[EtNameLOCATION] = EtLOCATION
	EtTypes[EtNameCHECK] = EtCHECK
	EtTypes[EtNameBIND] = EtBIND

	EtNames = make(map[int]string)
	EtNames[EtTCP] = EtNameTCP
	EtNames[EtDNS] = EtNameDNS
	EtNames[EtDNS6] = EtNameDNS6
	EtNames[EtLOCATION] = EtNameLOCATION
	EtNames[EtCHECK] = EtNameCHECK
	EtNames[EtBIND] = EtNameBIND
}

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

// ParseProxyStatus 识别ProxyStatus
func ParseProxyStatus(status string) (int, error) {
	status = strings.ToLower(status)
	switch status {
	case "smart":
		return ProxySMART, nil
	case "enable":
		return ProxyENABLE, nil
	default:
		return 0, errors.New("invalid proxy status: " + status)
	}
}

// FormatProxyStatus 格式化ProxyStatus
func FormatProxyStatus(status int) string {
	switch status {
	case ErrorProxyStatus:
		return "ERROR"
	case ProxyENABLE:
		return "ENABLE"
	case ProxySMART:
		return "SMART"
	default:
		return "UNKNOWN"
	}
}

// ParseEtType 得到字符串对应的ET请求类型
func ParseEtType(src string) int {
	src = strings.ToUpper(src)
	theType, ok := EtTypes[src]
	if !ok {
		return EtUNKNOWN
	}
	return theType
}

// FormatEtType 得到ET请求类型对应的字符串
func FormatEtType(src int) string {
	name, ok := EtNames[src]
	if !ok {
		return EtNameUNKNOWN
	}
	return name
}
