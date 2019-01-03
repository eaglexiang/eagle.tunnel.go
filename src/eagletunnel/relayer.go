/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-03 15:27:00
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-03 16:21:26
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
func (relayer *Relayer) Start() {
	var err error
	relayer.running = true

	// disable tls check for ip-inside cache
	http.DefaultTransport.(*http.Transport).TLSClientConfig =
		&tls.Config{InsecureSkipVerify: true}

	ipe := LocalAddr + ":" + LocalPort
	relayer.listener, err = net.Listen("tcp", ipe)
	if err != nil {
		fmt.Println("error: failed to listen! ", err)
	} else {
		fmt.Println("start to listen: ", ipe)
		relayer.listen()
	}
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
	tunnel := eaglelib.Tunnel{Left: &conn}
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
	if handler != nil {
		result := handler.Handle(request, &tunnel)
		if result {
			tunnel.Flow()
		} else {
			tunnel.Close()
		}
	} else {
		tunnel.Close()
	}
}

// SPrintConfig 将所有配置按格式输出为字符串
func SPrintConfig() string {
	var status string
	switch ProxyStatus {
	case ProxySMART:
		status = "smart"
	case ProxyENABLE:
		status = "enable"
	default:
	}
	var localUser string
	if LocalUser != nil {
		localUser = LocalUser.toString()
	} else {
		localUser = "null"
	}
	var userCheck string
	if EnableUserCheck {
		userCheck = "on"
	} else {
		userCheck = "off"
	}
	var http string
	if EnableHTTP {
		http = "on"
	} else {
		http = "off"
	}
	var socks string
	if EnableSOCKS5 {
		socks = "on"
	} else {
		socks = "off"
	}
	var ets string // et status
	if EnableET {
		ets = "on"
	} else {
		ets = "off"
	}

	var configStr string
	configStr += "config-path=" + ConfigPath + "\n"
	configStr += "config-dir=" + ConfigKeyValues["config-dir"] + "\n"
	configStr += "relayer=" + RemoteAddr + ":" + RemotePort + "\n"
	configStr += "listen=" + LocalAddr + ":" + LocalPort + "\n"
	configStr += "data-key=" + ConfigKeyValues["data-key"] + "\n"
	configStr += "user=" + localUser + "\n"
	configStr += "user-check=" + userCheck + "\n"
	configStr += "http=" + http + "\n"
	configStr += "socks=" + socks + "\n"
	configStr += "et=" + ets + "\n"
	configStr += "proxy-status=" + status + "\n"
	return configStr
}

// CheckSpeedOfUsers 轮询所有用户的速度，并根据配置选择是否进行限速
func CheckSpeedOfUsers() {
	for {
		for _, user := range Users {
			user.checkSpeed()
			user.limitSpeed()
		}

		LocalUser.checkSpeed()
		LocalUser.limitSpeed()

		time.Sleep(time.Second)
	}
}

// Close 关闭服务
func (relayer *Relayer) Close() {
	relayer.running = false
	if relayer.listener != nil {
		time.Sleep(time.Duration(1) * time.Second)
		relayer.listener.Close()
	}
}
