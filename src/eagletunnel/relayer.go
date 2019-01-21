/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-03 15:27:00
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-21 17:37:31
 */

package eagletunnel

import (
	"fmt"
	"net"

	mytunnel "github.com/eaglexiang/go-tunnel"
)

// LocalAddr 本地监听地址
var LocalAddr string

// LocalPort 本地监听端口
var LocalPort string

// LocalUser 本地用户
var LocalUser *EagleUser

// Users 需要鉴权的下级用户
var Users map[string]*EagleUser

// Relayer 网络入口，负责流量分发
type Relayer struct {
}

// Handle 处理请求连接
func (relayer *Relayer) Handle(conn net.Conn) {
	var buffer = make([]byte, 1024)
	count, err := conn.Read(buffer)
	if err != nil {
		return
	}
	request := Request{requestMsg: buffer[:count]}
	tunnel := mytunnel.GetTunnel()
	defer mytunnel.PutTunnel(tunnel)
	tunnel.Left = &conn
	var handler Handler
	switch request.getType() {
	case EagleTunnelReq:
		if EnableET {
			handler = new(EagleTunnel)
		}
	case HTTPProxyReq:
		if EnableHTTP {
			handler = new(HTTPProxy)
		}
	case SOCKSReq:
		if EnableSOCKS5 {
			handler = new(Socks5)
		}
	default:
		handler = nil
	}
	if handler == nil {
		return
	}
	err = handler.Handle(request, tunnel)
	if err != nil {
		if err.Error() != "no need to continue" {
			if ConfigKeyValues["debug"] == "on" {
				fmt.Println(err.Error())
			}
		}
		return
	}
	tunnel.Flow()
}
