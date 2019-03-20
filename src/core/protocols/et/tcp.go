/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-23 22:54:58
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-03-17 20:00:01
 */

package et

import (
	"errors"
	"net"
	"strings"

	"github.com/eaglexiang/eagle.tunnel.go/src/logger"
	mynet "github.com/eaglexiang/go-net"
	mytunnel "github.com/eaglexiang/go-tunnel"
)

// TCP ET-TCP子协议的实现
type TCP struct {
	arg  *Arg
	dns  DNS
	dns6 DNS
}

// Send 发送请求
func (t TCP) Send(et *ET, e *NetArg) (err error) {
	// 检查目的地址是否合法
	if e.IP == "" && e.Domain == "" {
		// 不存在可供使用的IP或域名
		return errors.New("TCP.Send -> no des host")
	}

	if e.IP == "" {
		// IP不存在，解析域名
		err = t.resolvDNS(et, e)
		if err != nil {
			return err
		}
	}

	// 建立连接
	switch t.arg.ProxyStatus {
	case ProxySMART:
		err = t.smartSend(et, e)
	case ProxyENABLE:
		err = t.proxySend(et, e)
	default:
		err = errors.New("invalid proxy-status")
	}

	if err != nil {
		return errors.New("TCP.Send -> ")
	}
	return nil
}

func (t TCP) resolvDNS(et *ET, e *NetArg) (err error) {
	// 调用DNS Sender解析Domain为IP
	switch t.arg.IPType {
	case "4":
		err = t.dns.Send(et, e)
	case "6":
		err = t.dns6.Send(et, e)
	case "46":
		err = t.dns.Send(et, e)
		if err != nil {
			err = t.dns6.Send(et, e)
		}
	case "64":
		err = t.dns6.Send(et, e)
		if err != nil {
			err = t.dns.Send(et, e)
		}
	default:
		logger.Warning("invalid ip-type: ", t.arg.IPType)
		err = errors.New("TCP.Send -> invalid ip-type")
	}
	return err
}

func (t *TCP) smartSend(et *ET, e *NetArg) (err error) {
	l := et.subSenders[EtLOCATION].(Location)
	err = l.Send(et, e)
	if err != nil {
		return err
	}
	if l.CheckProxyByLocation(e.Location) {
		err = t.sendTCPReq2Remote(et, e)
	} else {
		err = t.sendTCPReq2Server(e)
	}
	return err
}

func (t *TCP) proxySend(et *ET, e *NetArg) error {
	return t.sendTCPReq2Remote(et, e)
}

// Type ET子协议的类型
func (t TCP) Type() int {
	return EtTCP
}

func (t *TCP) sendTCPReq2Remote(et *ET, e *NetArg) error {
	err := et.connect2Relayer(e.Tunnel)
	if err != nil {
		return err
	}
	req := FormatEtType(EtTCP) + " " + e.IP + " " + e.Port
	_, err = e.Tunnel.WriteRight([]byte(req))
	if err != nil {
		return err
	}
	reply, err := e.Tunnel.ReadRightStr()
	if reply != "ok" {
		logger.Warning("invalid reply for ", req, ": ", reply)
		err = errors.New("failed 2 connect 2 server by relayer")
	}
	return err
}

func (t *TCP) sendTCPReq2Server(e *NetArg) error {
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
	conn, err := net.DialTimeout("tcp", ipe, t.arg.Timeout)
	if err != nil {
		logger.Warning(err)
		return err
	}
	e.Tunnel.Right = conn
	e.Tunnel.EncryptRight = nil
	e.Tunnel.DecryptRight = nil
	e.Tunnel.SpeedLimiter = t.arg.LocalUser.SpeedLimiter()
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
	e := NetArg{
		NetConnArg: NetConnArg{
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

// Match 判断是否匹配
func (t TCP) Match(req string) bool {
	args := strings.Split(req, " ")
	if args[0] == "TCP" {
		return true
	}
	return false
}
