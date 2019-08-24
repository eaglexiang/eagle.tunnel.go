/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-02-24 22:53:52
 * @LastEditTime: 2019-08-24 10:47:58
 */

package cmd

import (
	"errors"
	"strings"

	"github.com/eaglexiang/eagle.tunnel.go/core/protocols/et/comm"
	"github.com/eaglexiang/go-logger"
	"github.com/eaglexiang/go-tunnel"
)

// Bind BIND子协议
type Bind struct {
}

// Type 类型
func (b Bind) Type() comm.CMDType {
	return comm.BIND
}

// Name ET子协议的名字
func (b Bind) Name() string {
	return comm.BINDTxt
}

// Send 发送请求
func (b Bind) Send(e *comm.NetArg) error {
	// 连接远端
	err := comm.Connect2Remote(e.Tunnel)
	if err != nil {
		return err
	}
	// 发送请求
	req := comm.FormatEtType(comm.BIND) + " " + e.Port
	_, err = e.Tunnel.WriteRight([]byte(req))
	if err != nil {
		return err
	}
	reply, err := e.Tunnel.ReadRightStr()
	ipe := strings.Split(reply, " ")
	if len(ipe) != 2 {
		logger.Warning("invalid reply for bind.Send: ", reply)
		return errors.New("invalid reply")
	}
	e.IP = ipe[0]
	e.Port = ipe[1]
	return nil
}

// Handle 处理请求
func (b Bind) Handle(req string, tunnel *tunnel.Tunnel) (err error) {
	// reqs := strings.Split(req, " ")
	// if len(reqs) != 2 {
	// 	return errors.New("bind.Handle -> invalid req: " + req)
	// }
	// var listener net.Listener
	// if listener, err = net.Listen("tcp", "0.0.0.0:"+reqs[1]); err != nil {
	// 	reply := err.Error()
	// 	tunnel.WriteLeft([]byte(reply))
	// 	return errors.New("bind.Handle -> " + reply)
	// }
	// addrs, err := net.InterfaceAddrs()
	// var reply string
	// if err != nil {
	// 	reply = err.Error()
	// 	tunnel.WriteLeft([]byte(reply))
	// 	return errors.New("bind.Handle -> " + reply)
	// }
	// for _, addr := range addrs {
	// 	if ip, ok := addr.(*net.IPNet); ok && !ip.IP.IsLoopback() &&
	// 		!ip.IP.IsLinkLocalMulticast() && !ip.IP.IsLinkLocalUnicast() {
	// 		reply = ip.String()
	// 		tunnel.WriteLeft([]byte(reply))
	// 		conn, err := listener.Accept()
	// 		if err != nil {
	// 			return errors.New("bind.Handle -> " +
	// 				err.Error())
	// 		}
	// 	}
	// }
	// panic("no valid ip")
	return nil
}
