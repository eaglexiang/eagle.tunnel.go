/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-13 18:54:13
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-13 04:42:38
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
func (ed *ETDNS) Handle(req Request, tunnel *eaglelib.Tunnel) error {
	reqs := strings.Split(req.RequestMsgStr, " ")
	if len(reqs) < 2 {
		return errors.New("ETDNS.Handle -> req is too short")
	}
	domain := reqs[1]
	e := NetArg{domain: domain}
	err := resolvDNSByLocalServer(&e)
	if err != nil {
		return errors.New("ETDNS.Handle -> " + err.Error())
	}
	_, err = tunnel.WriteLeft([]byte(e.IP))
	if err != nil {
		return errors.New("ETDNS.Handle -> " + err.Error())
	}
	return nil
}

// Send 发送ET-DNS请求
func (ed *ETDNS) Send(e *NetArg) error {
	ip, result := hostsCache[e.domain]
	if result {
		e.IP = ip
		return nil
	}
	switch ProxyStatus {
	case ProxySMART:
		white := IsWhiteDomain(e.domain)
		if white {
			err := resolvDNSByProxy(e)
			if err != nil {
				return errors.New("ETDNS.Send -> " + err.Error())
			}
			return nil
		}
		err := resolvDNSByLocalClient(e)
		if err != nil {
			return errors.New("ETDNS.Send -> " + err.Error())
		}
		return nil
	case ProxyENABLE:
		err := resolvDNSByProxy(e)
		if err != nil {
			return errors.New("ETDNS.Send -> " + err.Error())
		}
		return nil
	default:
		return errors.New("ETDNS.Send -> invalid proxy-status")
	}
}

func resolvDNSByProxy(e *NetArg) error {
	var err error
	if dnsRemoteCache.Exsit(e.domain) {
		e.IP, err = dnsRemoteCache.Wait4IP(e.domain)
		if err != nil {
			return errors.New("resolvDNSByProxy -> " + err.Error())
		}
		return nil
	}
	dnsRemoteCache.Add(e.domain)
	err = _resolvDNSByProxy(e)
	if err != nil {
		dnsRemoteCache.Delete(e.domain)
		return errors.New("resolvDNSByProxy -> " + err.Error())
	}
	dnsRemoteCache.Update(e.domain, e.IP)
	return nil
}

func _resolvDNSByProxy(e *NetArg) error {
	// connect 2 relayer
	tunnel := eaglelib.GetTunnel()
	defer eaglelib.PutTunnel(tunnel)
	err := connect2Relayer(tunnel)
	if err != nil {
		return errors.New("_resolvDNSByProxy -> " + err.Error())
	}
	// send req
	req := FormatEtType(EtDNS) + " " + e.domain
	count, err := tunnel.WriteRight([]byte(req))
	if err != nil {
		return errors.New("_resolvDNSByProxy -> " + err.Error())
	}
	// get reply
	buffer := make([]byte, 1024)
	count, err = tunnel.ReadRight(buffer)
	if err != nil {
		return errors.New("_resolvDNSByProxy -> " + err.Error())
	}
	reply := string(buffer[:count])
	ip := net.ParseIP(reply)
	if ip == nil {
		return errors.New("_resolvDNSByProxy -> failed to resolv by remote: " +
			e.domain + " -> " + reply)
	}
	e.IP = reply
	return nil
}

// resolvDNSByLocalClient 本地解析DNS，如无ip-type配置，优先返回IPv4
func resolvDNSByLocalClient(e *NetArg) (err error) {
	if dnsLocalCache.Exsit(e.domain) {
		e.IP, err = dnsLocalCache.Wait4IP(e.domain)
		if err != nil {
			return errors.New("resolvDNSByLocalClient -> " + err.Error())
		}
		return nil
	}
	dnsLocalCache.Add(e.domain)
	err = _resolvDNSByLocalClient(e)
	if err != nil {
		dnsLocalCache.Delete(e.domain)
		return errors.New("resolvDNSByLocalClient -> " + err.Error())
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
		return errors.New("_resolvDNSByLocalClient -> fail to resolv DNS by local: " + e.domain)
	}

	// 判断IP所在位置是否适合代理
	el := ETLocation{}
	el.Send(e)
	proxy := CheckProxyByLocation(e.Reply)
	if proxy {
		// 更新IP为Relayer端的解析结果
		ne := NetArg{domain: e.domain}
		err = resolvDNSByProxy(&ne)
		if err != nil {
			return errors.New("_resolvDNSByLocalClient -> " + err.Error())
		}
		e.IP = ne.IP
	}

	return nil
}

// resolvDNSByLocalServer 本地解析DNS，如无ip-type配置，优先返回IPv4
func resolvDNSByLocalServer(e *NetArg) (err error) {
	if dnsLocalCache.Exsit(e.domain) {
		e.IP, err = dnsLocalCache.Wait4IP(e.domain)
		if err != nil {
			return errors.New("resolvDNSByLocalServer -> " + err.Error())
		}
		return nil
	}
	dnsLocalCache.Add(e.domain)
	err = _resolvDNSByLocalServer(e)
	if err != nil {
		dnsLocalCache.Delete(e.domain)
		return errors.New("resolvDNSByLocalServer -> " + err.Error())
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
	} else if ConfigKeyValues["ip-type"] == "46" {
		firstIPResolver = ResolvIPv4ByLocal
		secondIPResolver = ResolvIPv6ByLocal
	} else if ConfigKeyValues["ip-type"] == "64" {
		firstIPResolver = ResolvIPv6ByLocal
		secondIPResolver = ResolvIPv4ByLocal
	} else {
		panic("invalid ip-type: " + ConfigKeyValues["ip-type"])
	}

	err := firstIPResolver(e)
	if err != nil {
		err = secondIPResolver(e)
	}

	if err != nil {
		return errors.New("_resolvDNSByLocalServer -> " + err.Error())
	}
	return nil
}
