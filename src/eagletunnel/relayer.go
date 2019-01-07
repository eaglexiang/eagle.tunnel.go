/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-03 15:27:00
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-07 21:14:47
 */

package eagletunnel

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"../eaglelib/src"
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
	listener net.Listener
	running  bool
}

// Start 开始服务
func (relayer *Relayer) Start() error {
	var err error
	relayer.running = true

	// disable tls check for ip-inside cache
	http.DefaultTransport.(*http.Transport).TLSClientConfig =
		&tls.Config{InsecureSkipVerify: true}

	ipe := LocalAddr + ":" + LocalPort
	relayer.listener, err = net.Listen("tcp", ipe)
	if err != nil {
		return err
	}
	fmt.Println("start to listen: ", ipe)

	relayer.listen()
	return nil
}

func (relayer *Relayer) listen() {
	for relayer.running {
		conn, err := relayer.listener.Accept()
		if err != nil {
			fmt.Println("stop to accept! ", err)
			break
		} else {
			go relayer.handleClient(conn)
		}
		time.Sleep(time.Millisecond)
	}
	fmt.Println("quit")
}

func (relayer *Relayer) handleClient(conn net.Conn) {
	var buffer = make([]byte, 1024)
	count, err := conn.Read(buffer)
	if err != nil {
		return
	}
	request := Request{requestMsg: buffer[:count]}
	tunnel := eaglelib.GetTunnel()
	defer eaglelib.PutTunnel(tunnel)
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

// Close 关闭服务
func (relayer *Relayer) Close() {
	relayer.running = false
	if relayer.listener != nil {
		time.Sleep(time.Duration(1) * time.Second)
		relayer.listener.Close()
	}
}
