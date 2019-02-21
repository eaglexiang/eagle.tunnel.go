/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-23 22:54:58
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-02-21 18:23:51
 */

package et

import (
	"errors"
	"net"
	"strings"
	"time"

	"go.uber.org/ratelimit"

	mytunnel "github.com/eaglexiang/go-tunnel"
)

// TCP ET-TCP子协议的实现
// 必须使用createTCP进行构造
type TCP struct {
	proxyStatus int
	ipType      string
	limiter     *ratelimit.Limiter
	timeout     time.Duration
	dns         DNS
	dns6        DNS6
}

// createTCP 构造TCP
func createTCP(
	proxyStatus int,
	limiter *ratelimit.Limiter,
	timeout time.Duration,
	ipType string,
	dns DNS,
	dns6 DNS6,
) TCP {
	return TCP{
		proxyStatus: proxyStatus,
		limiter:     limiter,
		timeout:     timeout,
		ipType:      ipType,
		dns:         dns,
		dns6:        dns6,
	}
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
			return errors.New("TCP.Send -> " +
				err.Error())
		}
	}

	// 建立连接
	switch t.proxyStatus {
	case ProxySMART:
		err = t.smartSend(et, e)
	case ProxyENABLE:
		err = t.proxySend(et, e)
	default:
		err = errors.New("invalid proxy-status")
	}

	if err != nil {
		return errors.New("TCP.Send -> " +
			err.Error())
	}
	return nil
}

func (t TCP) resolvDNS(et *ET, e *NetArg) (err error) {
	// 调用DNS Sender解析Domain为IP
	switch t.ipType {
	case "4":
		err = t.dns.Send(et, e)
		if err != nil {
			return errors.New("TCP.Send -> " +
				err.Error())
		}
	case "6":
		err = t.dns6.Send(et, e)
		if err != nil {
			return errors.New("TCP.Send -> " +
				err.Error())
		}
	case "46":
		err = t.dns.Send(et, e)
		if err != nil {
			err = t.dns6.Send(et, e)
		}
		if err != nil {
			return errors.New("TCP.Send -> " +
				err.Error())
		}
	case "64":
		err = t.dns6.Send(et, e)
		if err != nil {
			err = t.dns.Send(et, e)
		}
		if err != nil {
			return errors.New("TCP.Send -> " +
				err.Error())
		}
	default:
		return errors.New("TCP.Send -> invalid ip-type: " +
			t.ipType)
	}
	return nil
}

func (t *TCP) smartSend(et *ET, e *NetArg) error {
	l := et.subSenders[EtLOCATION].(Location)
	l.Send(et, e)
	proxy := l.CheckProxyByLocation(e.Location)
	if proxy {
		// 启用代理
		err := t.sendTCPReq2Remote(et, e)
		if err != nil {
			return errors.New("TCP.smartSend -> " + err.Error())
		}
		return nil
	}
	// 不启用代理
	err := t.sendTCPReq2Server(e)
	if err != nil {
		// 直连失败的网站应被用户察觉
		return errors.New("TCP.smartSend -> " + err.Error())
	}
	return nil
}

func (t *TCP) proxySend(et *ET, e *NetArg) error {
	err := t.sendTCPReq2Remote(et, e)
	if err != nil {
		return errors.New("TCP.proxySend -> " + err.Error())
	}
	return nil
}

// Type ET子协议的类型
func (t TCP) Type() int {
	return EtTCP
}

func (t *TCP) sendTCPReq2Remote(et *ET, e *NetArg) error {
	req := FormatEtType(EtTCP) + " " + e.IP + " " + e.Port
	reply := sendQueryReq(et, req)
	if reply != "ok" {
		return errors.New("TCP.sendTCPReq2Remote -> " +
			"failed 2 connect 2 server by relayer")
	}
	return nil
}

func (t *TCP) sendTCPReq2Server(e *NetArg) error {
	if e.IP == "0.0.0.0" {
		return nil
	}
	if e.IP == "::" {
		return nil
	}

	var ipe string
	ip := net.ParseIP(e.IP)
	if ip.To4() != nil {
		ipe = e.IP + ":" + e.Port // ipv4:port
	} else {
		ipe = "[" + e.IP + "]:" + e.Port // [ipv6]:port
	}
	conn, err := net.DialTimeout("tcp", ipe, t.timeout)
	if err != nil {
		return errors.New("TCP.sendTCPReq2Server -> " + err.Error())
	}
	e.Tunnel.Right = &conn
	e.Tunnel.EncryptRight = nil
	e.Tunnel.DecryptRight = nil
	e.Tunnel.SpeedLimiter = t.limiter
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
	e := NetArg{IP: ip, Port: port, Tunnel: tunnel}
	err := t.sendTCPReq2Server(&e)
	if err != nil {
		tunnel.WriteLeft([]byte("nok"))
		return errors.New("TCP.Handle -> " + err.Error())

	}
	_, err = tunnel.WriteLeft([]byte("ok"))
	if err != nil {
		return errors.New("TCP.Handle -> " + err.Error())
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
