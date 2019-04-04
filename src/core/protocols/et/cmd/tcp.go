/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-23 22:54:58
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-04-03 20:44:20
 */

package cmd

import (
	"errors"
	"net"
	"strings"

	"github.com/eaglexiang/eagle.tunnel.go/src/core/protocols/et/comm"
	"github.com/eaglexiang/eagle.tunnel.go/src/logger"
	mynet "github.com/eaglexiang/go-net"
	"github.com/eaglexiang/go-tunnel"
)

var dnsResolver map[string]func(*comm.NetArg) error

func init() {
	dnsResolver = make(map[string]func(*comm.NetArg) error)
	dnsResolver["4"] = resolvDNS4
	dnsResolver["6"] = resolvDNS6
	dnsResolver["46"] = resolvDNS46
	dnsResolver["64"] = resolvDNS64
}

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

func resolvDNS4(e *comm.NetArg) error {
	return comm.SubSenders[comm.DNS].Send(e)
}

func resolvDNS6(e *comm.NetArg) error {
	return comm.SubSenders[comm.DNS6].Send(e)
}

func resolvDNS46(e *comm.NetArg) (err error) {
	if err = resolvDNS4(e); err != nil {
		err = resolvDNS6(e)
	}
	return
}

func resolvDNS64(e *comm.NetArg) (err error) {
	if err = resolvDNS6(e); err != nil {
		err = resolvDNS4(e)
	}
	return
}

func (t TCP) resolvDNS(e *comm.NetArg) error {
	resolver, ok := dnsResolver[comm.ETArg.IPType]
	if !ok {
		logger.Error("invalid ip-type", comm.ETArg.IPType)
		return errors.New("invalid ip-type")
	}
	return resolver(e)
}

func (t *TCP) smartSend(e *comm.NetArg) (err error) {
	switch e.DomainType {
	case comm.DirectDomain:
		logger.Info("connect direct domain ", e.Domain)
		err = t.connectByLocal(e)
	case comm.ProxyDomain:
		logger.Info("connect proxy domain ", e.Domain)
		err = t.connectByProxy(e)
	default:
		logger.Info("connect uncertain domain ", e.Domain)
		err = t.connectByLocation(e)
	}
	return err
}

func (t TCP) connectByLocation(e *comm.NetArg) (err error) {
	err = comm.SubSenders[comm.LOCATION].Send(e)
	if err != nil {
		return err
	}
	if checkProxyByLocation(e.Location) {
		err = t.connectByProxy(e)
	} else {
		err = t.connectByLocal(e)
	}
	return err
}

func (t TCP) proxySend(e *comm.NetArg) error {
	return t.connectByProxy(e)
}

// Type ET子协议的类型
func (t TCP) Type() comm.CMDType {
	return comm.TCP
}

// Name ET子协议的名字
func (t TCP) Name() string {
	return comm.TCPTxt
}

func (t *TCP) connectByProxy(e *comm.NetArg) error {
	err := comm.Connect2Remote(e.Tunnel)
	if err != nil {
		return err
	}
	logger.Info("connect ", e.IP+":"+e.Port, " by proxy")
	req := comm.FormatEtType(comm.TCP) + " " + e.IP + " " + e.Port
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

func (t *TCP) connectByLocal(e *comm.NetArg) error {
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
	logger.Info("connect ", ipe, " by local")
	conn, err := net.DialTimeout("tcp", ipe, comm.Timeout)
	if err != nil {
		logger.Warning(err)
		return err
	}
	e.Tunnel.Update(
		tunnel.WithRight(conn),
		tunnel.WithRightCipher(nil),
		tunnel.WithSpeedLimiter(comm.ETArg.LocalUser.SpeedLimiter()),
	)
	return nil
}

// Handle 处理ET-TCP请求
func (t TCP) Handle(req string, tn *tunnel.Tunnel) error {
	reqs := strings.Split(req, " ")
	if len(reqs) < 3 {
		return errors.New("TCP.Handle -> no des ip for tcp req")
	}
	ip := reqs[1]
	port := reqs[2]
	e := &comm.NetArg{
		NetConnArg: comm.NetConnArg{
			IP:   ip,
			Port: port,
		},
		Tunnel: tn,
	}
	err := t.connectByLocal(e)
	if err != nil {
		tn.WriteLeft([]byte(err.Error()))
		return err

	}
	_, err = tn.WriteLeft([]byte("ok"))
	if err != nil {
		return err
	}
	return nil
}
