/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-21 17:22:18
 * @LastEditTime: 2019-01-21 17:37:55
 */

package eagletunnel

import mytunnel "github.com/eaglexiang/go-tunnel"

// NetArg 服务内部需要的参数集合
type NetArg struct {
	domain  string
	IP      string
	Port    int
	tunnel  *mytunnel.Tunnel
	user    *EagleUser
	boolObj bool
	TheType int
	Reply   string
	Args    []string
}

// Clone 克隆一个NetArg，深拷贝Args字段
func (na *NetArg) Clone() *NetArg {
	result := NetArg{
		domain:  na.domain,
		IP:      na.IP,
		Port:    na.Port,
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
