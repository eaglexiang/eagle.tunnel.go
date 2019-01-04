/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-13 18:54:13
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-04 14:00:23
 */

package eagletunnel

import (
	"errors"
	"net"
	"strings"

	"../eaglelib/src"
)

var dnsRemoteCache = eaglelib.CreateDNSCache()
var dnsLocalCache = eaglelib.CreateDNSCache()
var hostsCache = make(map[string]string)

// ETDNS ET-DNS子协议的实现
type ETDNS struct {
}

// Handle 处理ET-DNS请求
func (ed *ETDNS) Handle(req Request, tunnel *eaglelib.Tunnel) {
	reqs := strings.Split(req.RequestMsgStr, " ")
	if len(reqs) >= 2 {
		domain := reqs[1]
		e := NetArg{domain: domain}
		err := resolvDNSByLocalServer(&e)
		if err == nil {
			tunnel.WriteLeft([]byte(e.IP))
		}
	}
}

// Send 发送ET-DNS请求
func (ed *ETDNS) Send(e *NetArg) bool {
	ip, result := hostsCache[e.domain]
	if result {
		e.IP = ip
		return true
	}
	switch ProxyStatus {
	case ProxySMART:
		white := IsWhiteDomain(e.domain)
		if white {
			result = resolvDNSByProxy(e) == nil
		} else {
			result = resolvDNSByLocalClient(e) == nil
		}
	case ProxyENABLE:
		result = resolvDNSByProxy(e) == nil
	default:
		result = false
	}
	return result
}

func resolvDNSByProxy(e *NetArg) error {
	var err error
	if dnsRemoteCache.Exsit(e.domain) {
		e.IP, err = dnsRemoteCache.Wait4IP(e.domain)
		return err
	}
	dnsRemoteCache.Add(e.domain)
	err = _resolvDNSByProxy(e)
	if err != nil {
		dnsRemoteCache.Delete(e.domain)
		return err
	}
	dnsRemoteCache.Update(e.domain, e.IP)
	return nil
}

func _resolvDNSByProxy(e *NetArg) error {
	tunnel := eaglelib.CreateTunnel()
	defer tunnel.Close()
	err := connect2Relayer(tunnel)
	if err != nil {
		return err
	}
	req := FormatEtType(EtDNS) + " " + e.domain
	count, err := tunnel.WriteRight([]byte(req))
	if err != nil {
		return err
	}
	buffer := make([]byte, 1024)
	count, err = tunnel.ReadRight(buffer)
	if err != nil {
		return err
	}
	_ip := string(buffer[:count])
	ip := net.ParseIP(_ip)
	if ip == nil {
		return errors.New("failed to resolv by remote")
	}
	e.IP = _ip
	return err
}

// resolvDNSByLocalClient 本地解析DNS，如无ip-type配置，优先返回IPv4
func resolvDNSByLocalClient(e *NetArg) (err error) {
	if dnsLocalCache.Exsit(e.domain) {
		e.IP, err = dnsLocalCache.Wait4IP(e.domain)
		return err
	}
	dnsLocalCache.Add(e.domain)
	err = _resolvDNSByLocalClient(e)
	if err != nil {
		dnsLocalCache.Delete(e.domain)
		return err
	}
	dnsLocalCache.Update(e.domain, e.IP)
	return nil
}

func _resolvDNSByLocalClient(e *NetArg) error {
	var firstIPResolver func(*NetArg) error
	var secondIPResolver func(*NetArg) error
	if ConfigKeyValues["ip-type"] == "4" {
		firstIPResolver = ResolvIPv4ByLocal
		secondIPResolver = ResolvIPv4ByLocal
	} else if ConfigKeyValues["ip-type"] == "6" {
		firstIPResolver = ResolvIPv6ByLocal
		secondIPResolver = ResolvIPv6ByLocal
	} else {
		firstIPResolver = ResolvIPv4ByLocal
		secondIPResolver = ResolvIPv6ByLocal
	}

	err := firstIPResolver(e)
	if err != nil {
		err = secondIPResolver(e)
	}

	// 本地解析失败应该让用户察觉，手动添加DNS白名单
	if err != nil {
		return errors.New("fail to resolv DNS by local")
	}

	// 判断IP所在位置是否适合代理
	el := ETLocation{}
	el.Send(e)
	proxy := CheckProxyByLocation(e.Reply)
	if proxy {
		// 更新IP为Relayer端的解析结果
		ne := NetArg{domain: e.domain}
		err = resolvDNSByProxy(&ne)
		if err == nil {
			e.IP = ne.IP
		}
	}

	return nil
}

// resolvDNSByLocalServer 本地解析DNS，如无ip-type配置，优先返回IPv4
func resolvDNSByLocalServer(e *NetArg) (err error) {
	if dnsLocalCache.Exsit(e.domain) {
		e.IP, err = dnsLocalCache.Wait4IP(e.domain)
		return err
	}
	dnsLocalCache.Add(e.domain)
	err = _resolvDNSByLocalServer(e)
	if err != nil {
		dnsLocalCache.Delete(e.domain)
		return err
	}
	dnsLocalCache.Update(e.domain, e.IP)
	return nil
}

func _resolvDNSByLocalServer(e *NetArg) error {
	var firstIPResolver func(*NetArg) error
	var secondIPResolver func(*NetArg) error
	if ConfigKeyValues["ip-type"] == "4" {
		firstIPResolver = ResolvIPv4ByLocal
		secondIPResolver = ResolvIPv4ByLocal
	} else if ConfigKeyValues["ip-type"] == "6" {
		firstIPResolver = ResolvIPv6ByLocal
		secondIPResolver = ResolvIPv6ByLocal
	} else {
		firstIPResolver = ResolvIPv4ByLocal
		secondIPResolver = ResolvIPv6ByLocal
	}

	err := firstIPResolver(e)
	if err != nil {
		err = secondIPResolver(e)
	}

	return err
}
