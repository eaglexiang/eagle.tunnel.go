/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-23 22:54:58
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-03-27 17:07:38
 */

package cmd

import (
	"errors"
	"net"
	"strings"

	"github.com/eaglexiang/eagle.tunnel.go/src/core/protocols/et/comm"
	"github.com/eaglexiang/eagle.tunnel.go/src/logger"
	mynet "github.com/eaglexiang/go-net"
	mytunnel "github.com/eaglexiang/go-tunnel"
)

// TCP ET-TCP子协议的实现
type TCP struct {
}

// Send 发送请求
func (t TCP) Send(e *comm.NetArg) (err error) {
	// 检查目的地址是否合法
	if e.IP == "" && e.Domain == "" {
		// 不存在可供使用的IP或域名
		return errors.New("TCP.Send -> no des host")
	}

	e.DomainType = comm.TypeOfDomain(e.Domain)

	if e.IP == "" {
		// IP不存在，解析域名
		err = t.resolvDNS(e)
		if err != nil {
			return err
		}
	}

	// 建立连接
	switch comm.ETArg.ProxyStatus {
	case comm.ProxySMART:
		err = t.smartSend(e)
	case comm.ProxyENABLE:
		err = t.proxySend(e)
	default:
		err = errors.New("invalid proxy-status")
	}

	if err != nil {
		return errors.New("TCP.Send -> ")
	}
	return nil
}

func (t TCP) resolvDNS(e *comm.NetArg) (err error) {
	// 调用DNS Sender解析Domain为IP
	switch comm.ETArg.IPType {
	case "4":
		err = comm.SubSenders[comm.EtDNS].Send(e)
	case "6":
		err = comm.SubSenders[comm.EtDNS6].Send(e)
	case "46":
		err = comm.SubSenders[comm.EtDNS].Send(e)
		if err != nil {
			err = comm.SubSenders[comm.EtDNS6].Send(e)
		}
	case "64":
		err = comm.SubSenders[comm.EtDNS6].Send(e)
		if err != nil {
			err = comm.SubSenders[comm.EtDNS].Send(e)
		}
	default:
		logger.Warning("invalid ip-type: ", comm.ETArg.IPType)
		err = errors.New("TCP.Send -> invalid ip-type")
	}
	return err
}

func (t *TCP) smartSend(e *comm.NetArg) (err error) {
	switch e.DomainType {
	case comm.DirectDomain:
		logger.Info("connect direct domain ", e.Domain)
		err = t.sendTCPReq2Server(e)
	case comm.ProxyDomain:
		logger.Info("connect proxy domain ", e.Domain)
		err = t.sendTCPReq2Remote(e)
	default:
		logger.Info("connect uncertain domain ", e.Domain)
		err = t.sendTCPReqByLocation(e)
	}
	return err
}

func (t TCP) sendTCPReqByLocation(e *comm.NetArg) (err error) {
	l := comm.SubSenders[comm.EtLOCATION]
	err = l.Send(e)
	if err != nil {
		return err
	}
	if checkProxyByLocation(e.Location) {
		err = t.sendTCPReq2Remote(e)
	} else {
		err = t.sendTCPReq2Server(e)
	}
	return err
}

func (t TCP) proxySend(e *comm.NetArg) error {
	return t.sendTCPReq2Remote(e)
}

// Type ET子协议的类型
func (t TCP) Type() int {
	return comm.EtTCP
}

// Name ET子协议的名字
func (t TCP) Name() string {
	return comm.EtNameTCP
}

func (t *TCP) sendTCPReq2Remote(e *comm.NetArg) error {
	err := comm.Connect2Remote(e.Tunnel)
	if err != nil {
		return err
	}
	req := comm.FormatEtType(comm.EtTCP) + " " + e.IP + " " + e.Port
	_, err = e.Tunnel.WriteRight([]byte(req))
	if err != nil {
		return err
	}
	reply, err := e.Tunnel.ReadRightStr()
	if err != nil {
		return err
	}
	if reply != "ok" {
		logger.Warning("invalid reply for ", req, ": ", reply)
		err = errors.New("failed 2 connect 2 server by relayer")
	}
	return err
}

func (t *TCP) sendTCPReq2Server(e *comm.NetArg) error {
	if e.IP == "0.0.0.0" || e.IP == "::" {
		logger.Info("invalid ip: ", e.IP)
		return errors.New("invalid ip")
	}

	var ipe string
	if mynet.TypeOfAddr(e.IP) == mynet.IPv4Addr {
		ipe = e.IP + ":" + e.Port // ipv4:port
	} else {
		ipe = "[" + e.IP + "]:" + e.Port // [ipv6]:port
	}
	conn, err := net.DialTimeout("tcp", ipe, comm.Timeout)
	if err != nil {
		logger.Warning(err)
		return err
	}
	e.Tunnel.Right = conn
	e.Tunnel.EncryptRight = nil
	e.Tunnel.DecryptRight = nil
	e.Tunnel.SpeedLimiter = comm.ETArg.LocalUser.SpeedLimiter()
	return nil
}

// Handle 处理ET-TCP请求
func (t TCP) Handle(req string, tunnel *mytunnel.Tunnel) error {
	reqs := strings.Split(req, " ")
	if len(reqs) < 3 {
		return errors.New("TCP.Handle -> no des ip for tcp req")
	}
	ip := reqs[1]
	port := reqs[2]
	e := comm.NetArg{
		NetConnArg: comm.NetConnArg{
			IP:   ip,
			Port: port,
		},
		Tunnel: tunnel,
	}
	err := t.sendTCPReq2Server(&e)
	if err != nil {
		tunnel.WriteLeft([]byte(err.Error()))
		return err

	}
	_, err = tunnel.WriteLeft([]byte("ok"))
	if err != nil {
		return err
	}
	return nil
}
