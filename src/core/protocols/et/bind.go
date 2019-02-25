/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-02-24 22:53:52
 * @LastEditTime: 2019-02-25 04:00:42
 */

package et

import (
	"errors"
	"strings"

	"github.com/eaglexiang/go-tunnel"
)

type bind struct {
}

// Match 判断是否匹配
func (b bind) Match(req string) bool {
	reqs := strings.Split(req, " ")
	if reqs[0] == "BIND" {
		return true
	}
	return false
}

// Type 类型
func (b bind) Type() int {
	return EtBIND
}

func (b bind) Send(et *ET, e *NetArg) error {
	// 连接远端
	err := et.connect2Relayer(e.Tunnel)
	if err != nil {
		return errors.New("bind.Send ->" +
			err.Error())
	}
	// 发送请求
	req := FormatEtType(EtBIND) + " " + e.Port
	_, err = e.Tunnel.WriteRight([]byte(req))
	if err != nil {
		return errors.New("bind.Send -> " +
			err.Error())
	}
	reply, err := e.Tunnel.ReadRightStr()
	ipe := strings.Split(reply, " ")
	if len(ipe) != 2 {
		return errors.New("bind.Send -> invalid reply: " +
			reply)
	}
	e.IP = ipe[0]
	e.Port = ipe[1]
	return nil
}

func (b bind) Handle(req string, tunnel *tunnel.Tunnel) (err error) {
	reqs := strings.Split(req, " ")
	if len(reqs) != 2 {
		return errors.New("bind.Handle -> invalid req: " + req)
	}
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
