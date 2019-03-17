/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-22 10:28:46
 * @LastEditTime: 2019-03-17 16:13:45
 */

package et

import (
	"errors"
	"net"
	"strings"

	mynet "github.com/eaglexiang/go-net"
	mytunnel "github.com/eaglexiang/go-tunnel"
)

// ET请求的类型
const (
	EtUNKNOWN = iota
	EtTCP
	EtDNS
	EtDNS6
	EtLOCATION
	EtCHECK
	EtBIND
)

// 代理的状态
const (
	ProxyENABLE = iota
	ProxySMART
	ErrorProxyStatus
)

// NetArg ET协议工作需要的参数集
// 此参数集用于在协议间传递消息
type NetArg struct {
	TheType      int
	Domain       string
	IP           string
	Port         string // 端口号
	Location     string // 所在地，用于识别是否使用代理
	BindDelegate func() // BIND操作会用到的委托
	Tunnel       *mytunnel.Tunnel
}

// 将net网络操作类型转化为ET网络操作类型
// 此函数供sender使用
func netOPType2ETOPType(netOPType int) int {
	switch netOPType {
	case mynet.CONNECT:
		return EtTCP
	case mynet.BIND:
		return EtBIND
	default:
		return EtUNKNOWN
	}
}

// parseNetArg 将通用的net.Arg转化为ET专用NetArg
func parseNetArg(e *mynet.Arg) (*NetArg, error) {
	ne := NetArg{
		TheType: netOPType2ETOPType(e.TheType),
		Tunnel:  e.Tunnel,
	}
	ipe := strings.Split(e.Host, ":")
	ne.Port = ipe[len(ipe)-1]
	_ip := strings.TrimSuffix(e.Host, ":"+ne.Port)
	ip := net.ParseIP(_ip)
	if ip != nil {
		ne.IP = _ip
	} else {
		ne.Domain = _ip
	}
	return &ne, nil
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
	switch src {
	case "DNS":
		return EtDNS
	case "DNS6":
		return EtDNS6
	case "TCP":
		return EtTCP
	case "LOCATION":
		return EtLOCATION
	case "CHECK":
		return EtCHECK
	case "BIND":
		return EtBIND
	default:
		return EtUNKNOWN
	}
}

// FormatEtType 得到ET请求类型对应的字符串
func FormatEtType(src int) string {
	switch src {
	case EtDNS:
		return "DNS"
	case EtDNS6:
		return "DNS6"
	case EtTCP:
		return "TCP"
	case EtLOCATION:
		return "LOCATION"
	case EtCHECK:
		return "CHECK"
	case EtBIND:
		return "BIND"
	default:
		return "UNKNOWN"
	}
}
