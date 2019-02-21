/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-22 10:28:46
 * @LastEditTime: 2019-02-21 18:41:47
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
	TheType  int
	Domain   string
	IP       string
	Port     string
	Location string
	Tunnel   *mytunnel.Tunnel
}

// parseNetArg 将通用的net.Arg转化为ET专用NetArg
func parseNetArg(e *mynet.Arg) (*NetArg, error) {
	ne := NetArg{
		TheType: e.TheType,
		Tunnel:  e.Tunnel,
	}
	ipe := strings.Split(e.Host, ":")
	if len(ipe) != 2 {
		return nil, errors.New("parseNetArg -> invalid Host: " +
			e.Host)
	}
	_ip := ipe[0]
	_port := ipe[1]
	ip := net.ParseIP(_ip)
	if ip != nil {
		ne.IP = _ip
	} else {
		ne.Domain = _ip
	}
	ne.Port = _port
	return &ne, nil
}

// ParseProxyStatus 识别ProxyStatus
func ParseProxyStatus(status string) int {
	switch status {
	case "smart", "Smart", "SMART":
		return ProxySMART
	case "enable", "Enable", "ENABLE":
		return ProxyENABLE
	default:
		return ErrorProxyStatus
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
	default:
		return "UNKNOWN"
	}
}
