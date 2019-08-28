/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-02-21 18:37:43
 * @LastEditTime: 2019-08-28 19:42:18
 */

package comm

import (
	"net"
	"strings"

	mynet "github.com/eaglexiang/go/net"
	"github.com/eaglexiang/go/tunnel"
	"github.com/eaglexiang/go/user"
)

// Arg 启动ET协议需要的的参数集
// 此参数集用于启动和配置ET协议服务
type Arg struct {
	ConnArg
	UsersArg
	SmartArg
	IPType string `label:"4/6/46/64 采用什么DNS解析模式"`
}

// UsersArg 用户系统使用的参数集
type UsersArg struct {
	LocalUser  *user.ValidUser
	ValidUsers map[string]*user.ValidUser
}

// ConnArg 连接相关参数集
type ConnArg struct {
	RemoteIPE string // IP:Port
	LocalIPEs []string
	Head      string // 协议头
}

// SmartArg 智能模式需要的参数集
type SmartArg struct {
	ProxyStatus   int
	LocalLocation string
}

// NetArg ET协议工作需要的参数集
// 此参数集用于在协议间传递消息
type NetArg struct {
	NetConnArg
	NetSmartArg
	TheType      int    // 网络请求的类型
	BindDelegate func() // BIND操作会用到的委托
	Tunnel       *tunnel.Tunnel
}

// NetSmartArg 与智能分流相关的参数
type NetSmartArg struct {
	Location   string // 所在地，用于识别是否使用代理
	DomainType int    // 域名的类型（强制代理/强制直连/不确定）
}

// NetConnArg NetArg中关于连接的部分
type NetConnArg struct {
	Domain string
	IP     string
	Port   string
}

// NetOPType2ETOPType 将net网络操作类型转化为ET网络操作类型
// 此函数供sender使用
func NetOPType2ETOPType(netOPType mynet.OpType) CMDType {
	switch netOPType {
	case mynet.CONNECT:
		return TCP
	case mynet.BIND:
		return BIND
	default:
		return UNKNOWN
	}
}

// ParseNetArg 将通用的net.Arg转化为ET专用NetArg
func ParseNetArg(e *mynet.Arg) (*NetArg, error) {
	ne := NetArg{
		TheType: int(NetOPType2ETOPType(mynet.OpType(e.TheType))),
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

// DefaultArg 运行ET协议所需的公共参数集
var DefaultArg *Arg
