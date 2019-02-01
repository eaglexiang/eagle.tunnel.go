/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-23 22:54:58
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-22 19:34:00
 */

package et

import (
	"errors"
	"net"
	"strings"
	"time"

	"github.com/eaglexiang/go-bytebuffer"

	mytunnel "github.com/eaglexiang/go-tunnel"
	myuser "github.com/eaglexiang/go-user"
)

// TCP ET-TCP子协议的实现 必须使用createTCP进行构造
type TCP struct {
	proxyStatus int
	localUser   *myuser.User
	timeout     time.Duration
}

// createTCP 构造TCP
func createTCP(
	proxyStatus int,
	localUser *myuser.User,
	timeout time.Duration,
) TCP {
	t := TCP{
		proxyStatus: proxyStatus,
		localUser:   localUser,
		timeout:     timeout,
	}
	return t
}

// Send 发送请求
func (t TCP) Send(et *ET, e *NetArg) (err error) {
	// 检查目的地址是否合法
	if e.IP == "" {
		if e.Domain == "" {
			return errors.New("TCP.Send -> no des host")
		}
		// 调用DNS解析Domain为IP
		_dns, ok := et.subSenders[EtDNS]
		if !ok {
			return errors.New("TCP.Send -> no dns sender")
		}
		dns := _dns.(DNS)
		err = dns.Send(et, e)
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

// ETType ET子协议的类型
func (t TCP) ETType() int {
	return EtTCP
}

func (t *TCP) sendTCPReq2Remote(et *ET, e *NetArg) error {
	err := et.connect2Relayer(e.Tunnel)
	if err != nil {
		return errors.New("TCP.sendTCPReq2Remote -> " + err.Error())
	}
	req := FormatEtType(EtTCP) + " " + e.IP + " " + e.Port
	_, err = e.Tunnel.WriteRight([]byte(req))
	if err != nil {
		return errors.New("TCP.sendTCPReq2Remote -> " + err.Error())
	}
	buffer := bytebuffer.GetKBBuffer()
	defer bytebuffer.PutKBBuffer(buffer)
	buffer.Length, err = e.Tunnel.ReadRight(buffer.Buf())
	if err != nil {
		return errors.New("TCP.sendTCPReq2Remote -> " + err.Error())
	}
	reply := buffer.String()
	if reply != "ok" {
		err = errors.New("TCP.sendTCPReq2Remote -> failed 2 connect 2 server by relayer")
	}
	return err
}

func (t *TCP) sendTCPReq2Server(e *NetArg) error {
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
	e.Tunnel.SpeedLimiter = t.localUser.SpeedLimiter()
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
