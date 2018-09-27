package eagletunnel

import (
	"github.com/eaglexiang/eagle.lib.go/src"
)

// NetArg 服务内部需要的参数集合
type NetArg struct {
	domain  string
	ip      string
	port    int
	tunnel  *eaglelib.Tunnel
	user    *EagleUser
	boolObj bool
	TheType int
	Reply   string
	Args    []string
}

// Clone 以深拷贝的方式克隆一个NetArg
func (na *NetArg) Clone() *NetArg {
	result := NetArg{
		domain:  na.domain,
		ip:      na.ip,
		port:    na.port,
		tunnel:  na.tunnel,
		user:    na.user,
		boolObj: na.boolObj,
		TheType: na.TheType,
		Reply:   na.Reply,
	}
	result.Args = make([]string, len(na.Args))
	for ind, item := range na.Args {
		result.Args[ind] = item
	}
	return &result
}
