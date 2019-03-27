/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-03-24 22:35:45
 * @LastEditTime: 2019-03-27 22:48:21
 */
package comm

import (
	"errors"
	"strings"
	"time"

	"github.com/eaglexiang/go-trie"
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

// 域名的类型
const (
	UncertainDomain = iota
	ProxyDomain
	DirectDomain
)

// 代理状态对应的文本
const (
	ProxyEnableText      = "ENABLE"
	ProxySmartText       = "SMART"
	ErrorProxyStatusText = "ERROR"
)

// EtTypes ET子协议的类型
var EtTypes map[string]int

// EtNames ET子协议的名字
var EtNames map[int]string

// EtProxyStatus ET代理状态
var EtProxyStatus map[string]int

// EtProxyStatusText ET代理状态对应的文本
var EtProxyStatusText map[int]string

// HostsCache 本地Hosts
var HostsCache = make(map[string]string)

// ProxyDomains 强制代理的域名列表
var ProxyDomains trie.StringTrie

// DirectDomains 强制直连的域名列表
var DirectDomains trie.StringTrie

// Timeout 超时长度
var Timeout time.Duration

func init() {
	EtTypes = make(map[string]int)
	EtTypes[EtNameTCP] = EtTCP
	EtTypes[EtNameDNS] = EtDNS
	EtTypes[EtNameDNS6] = EtDNS6
	EtTypes[EtNameLOCATION] = EtLOCATION
	EtTypes[EtNameCHECK] = EtCHECK
	EtTypes[EtNameBIND] = EtBIND
}

func init() {
	EtNames = make(map[int]string)
	EtNames[EtTCP] = EtNameTCP
	EtNames[EtDNS] = EtNameDNS
	EtNames[EtDNS6] = EtNameDNS6
	EtNames[EtLOCATION] = EtNameLOCATION
	EtNames[EtCHECK] = EtNameCHECK
	EtNames[EtBIND] = EtNameBIND
}

func init() {
	EtProxyStatus = make(map[string]int)
	EtProxyStatus[ProxyEnableText] = ProxyENABLE
	EtProxyStatus[ProxySmartText] = ProxySMART

	EtProxyStatusText = make(map[int]string)
	EtProxyStatusText[ProxyENABLE] = ProxyEnableText
	EtProxyStatusText[ProxySMART] = ProxySmartText
}

// ParseProxyStatus 识别ProxyStatus
func ParseProxyStatus(status string) (int, error) {
	status = strings.ToUpper(status)
	s, ok := EtProxyStatus[status]
	if !ok {
		return ErrorProxyStatus, errors.New(ErrorProxyStatusText)
	}
	return s, nil
}

// FormatProxyStatus 格式化ProxyStatus
func FormatProxyStatus(status int) string {
	s, ok := EtProxyStatusText[status]
	if !ok {
		return ErrorProxyStatusText
	}
	return s
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

// TypeOfDomain 域名的类型（强制代理/强制直连/不确定）
func TypeOfDomain(domain string) (status int) {
	if ProxyDomains.MatchSuffix(domain) {
		return ProxyDomain
	}
	if DirectDomains.MatchSuffix(domain) {
		return DirectDomain
	}
	return UncertainDomain
}
