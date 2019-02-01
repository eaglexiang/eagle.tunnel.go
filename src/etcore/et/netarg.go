/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-22 10:28:46
 * @LastEditTime: 2019-01-22 20:47:21
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
	EtLOCATION
	EtCHECK
)

// 代理的状态
const (
	ErrorProxyStatus = iota
	ProxyENABLE
	ProxySMART
)

// NetArg ET协议工作需要的参数集
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
	var result int
	switch src {
	case "DNS":
		result = EtDNS
	case "TCP":
		result = EtTCP
	case "LOCATION":
		result = EtLOCATION
	case "CHECK":
		result = EtCHECK
	default:
		result = EtUNKNOWN
	}
	return result
}

// FormatEtType 得到ET请求类型对应的字符串
func FormatEtType(src int) string {
	var result string
	switch src {
	case EtDNS:
		result = "DNS"
	case EtTCP:
		result = "TCP"
	case EtLOCATION:
		result = "LOCATION"
	case EtCHECK:
		result = "CHECK"
	default:
		result = "UNKNOWN"
	}
	return result
}
