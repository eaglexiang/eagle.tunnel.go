/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-02-21 18:37:43
 * @LastEditTime: 2019-02-21 23:08:25
 */

package et

import (
	"net"
	"strings"
	"time"

	mynet "github.com/eaglexiang/go-net"
	mytunnel "github.com/eaglexiang/go-tunnel"
	myuser "github.com/eaglexiang/go-user"
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
	LocalUser  *myuser.ValidUser
	ValidUsers map[string]*myuser.ValidUser
}

// ConnArg 连接相关参数集
type ConnArg struct {
	RemoteIPE string // IP:Port
	Timeout   time.Duration
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
	TheType      int
	Location     string // 所在地，用于识别是否使用代理
	BindDelegate func() // BIND操作会用到的委托
	Tunnel       *mytunnel.Tunnel
}

// NetConnArg NetArg中关于连接的部分
type NetConnArg struct {
	Domain string
	IP     string
	Port   string
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
