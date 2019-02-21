/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-13 18:54:13
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-02-21 18:20:44
 */

package et

import (
	"errors"
	"net"
	"strings"

	dnscache "github.com/eaglexiang/go-dnscache"
	mynet "github.com/eaglexiang/go-net"
	mytunnel "github.com/eaglexiang/go-tunnel"
)

var dnsRemoteCache = dnscache.CreateDNSCache()
var dnsLocalCache = dnscache.CreateDNSCache()

// HostsCache 本地Hosts
var HostsCache = make(map[string]string)

// WhitelistDomains 需要被智能解析的DNS域名列表
var WhitelistDomains []string

// DNS ET-DNS子协议的实现
type DNS struct {
	ProxyStatus int
}

// Handle 处理ET-DNS请求
func (d DNS) Handle(req string, tunnel *mytunnel.Tunnel) error {
	reqs := strings.Split(req, " ")
	if len(reqs) < 2 {
		return errors.New("ETDNS.Handle -> req is too short")
	}
	domain := reqs[1]
	e := NetArg{Domain: domain}
	err := d.resolvDNSByLocalServer(&e)
	if err != nil {
		return errors.New("ETDNS.Handle -> " + err.Error())
	}
	_, err = tunnel.WriteLeft([]byte(e.IP))
	if err != nil {
		return errors.New("ETDNS.Handle -> " + err.Error())
	}
	return nil
}

// Match 判断是否匹配
func (d DNS) Match(req string) bool {
	args := strings.Split(req, " ")
	if args[0] == "DNS" {
		return true
	}
	return false
}

// Send 发送ET-DNS请求
func (d DNS) Send(et *ET, e *NetArg) (err error) {
	ip, result := HostsCache[e.Domain]
	if result {
		e.IP = ip
		return nil
	}
	switch d.ProxyStatus {
	case ProxySMART:
		err = d.smartSend(et, e)
		if err != nil {
			return errors.New("DNS.Send -> " +
				err.Error())
		}
	case ProxyENABLE:
		err = d.proxySend(et, e)
		if err != nil {
			return errors.New("DNS.Send -> " +
				err.Error())
		}
	default:
		err = errors.New("DNS.Send -> invalid proxy-status")
	}

	if err != nil {
		return errors.New("DNS.Send -> " +
			err.Error())
	}
	return nil
}

// smartSend 智能模式
// 智能模式会先检查域名是否存在于白名单
// 白名单内域名将转入强制代理模式
func (d DNS) smartSend(et *ET, e *NetArg) error {
	white := IsWhiteDomain(e.Domain)
	if white {
		err := d.resolvDNSByProxy(et, e)
		if err != nil {
			return errors.New("DNS.smartSend -> " + err.Error())
		}
		return nil
	}
	err := d.resolvDNSByLocalClient(et, e)
	if err != nil {
		return errors.New("DNS.smartSend -> " + err.Error())
	}
	return nil
}

// proxySend 强制代理模式
func (d DNS) proxySend(et *ET, e *NetArg) error {
	err := d.resolvDNSByProxy(et, e)
	if err != nil {
		return errors.New("DNS.proxySend -> " + err.Error())
	}
	return nil
}

// Type ET子协议类型
func (d DNS) Type() int {
	return EtDNS
}

// resolvDNSByProxy 使用代理服务器进行DNS的解析
// 此函数主要完成缓存功能
// 当缓存不命中则调用 DNS._resolvDNSByProxy
func (d DNS) resolvDNSByProxy(et *ET, e *NetArg) error {
	var err error
	if dnsRemoteCache.Exsit(e.Domain) {
		e.IP, err = dnsRemoteCache.Wait4IP(e.Domain)
		if err != nil {
			return errors.New("resolvDNSByProxy -> " + err.Error())
		}
		return nil
	}
	dnsRemoteCache.Add(e.Domain)
	err = d._resolvDNSByProxy(et, e)
	if err != nil {
		dnsRemoteCache.Delete(e.Domain)
		return errors.New("resolvDNSByProxy -> " + err.Error())
	}
	dnsRemoteCache.Update(e.Domain, e.IP)
	return nil
}

// _resolvDNSByProxy 使用代理服务器进行DNS的解析
// 实际完成DNS查询操作
func (d DNS) _resolvDNSByProxy(et *ET, e *NetArg) error {
	req := FormatEtType(EtDNS) + " " + e.Domain
	reply := sendQueryReq(et, req)
	ip := net.ParseIP(reply)
	if ip == nil {
		return errors.New("_resolvDNSByProxy -> failed to resolv by remote: " +
			e.Domain + " -> " + reply)
	}
	e.IP = reply
	return nil
}

// resolvDNSByLocalClient 本地解析DNS
// 此函数由客户端使用
// 此函数主要完成缓存功能
// 当缓存不命中则进一步调用 DNS._resolvDNSByLocalClient
func (d DNS) resolvDNSByLocalClient(et *ET, e *NetArg) (err error) {
	if dnsLocalCache.Exsit(e.Domain) {
		e.IP, err = dnsLocalCache.Wait4IP(e.Domain)
		if err != nil {
			return errors.New("resolvDNSByLocalClient -> " + err.Error())
		}
		return nil
	}
	dnsLocalCache.Add(e.Domain)
	err = d._resolvDNSByLocalClient(et, e)
	if err != nil {
		dnsLocalCache.Delete(e.Domain)
		return errors.New("resolvDNSByLocalClient -> " + err.Error())
	}
	dnsLocalCache.Update(e.Domain, e.IP)
	return nil
}

// _resolvDNSByLocalClient 本地解析DNS
// 实际完成DNS的解析动作
func (d DNS) _resolvDNSByLocalClient(et *ET, e *NetArg) (err error) {
	e.IP, err = mynet.ResolvIPv4(e.Domain)
	// 本地解析失败应该让用户察觉，手动添加DNS白名单
	if err != nil {
		return errors.New("_resolvDNSByLocalClient -> fail to resolv DNS by local: " + e.Domain +
			" , consider adding this domain to your whitelist_domain.txt")
	}

	// 判断IP所在位置是否适合代理
	l := et.subSenders[EtLOCATION].(Location)
	l.Send(et, e)
	proxy := l.CheckProxyByLocation(e.Location)
	if proxy {
		// 更新IP为Relayer端的解析结果
		ne := NetArg{Domain: e.Domain}
		err = d.resolvDNSByProxy(et, &ne)
		if err != nil {
			return errors.New("_resolvDNSByLocalClient -> " + err.Error())
		}
		e.IP = ne.IP
	}

	return nil
}

// resolvDNSByLocalServer 本地解析DNS
// 此函数由服务端使用
// 此函数完成缓存相关的工作
// 当缓存不命中则进一步调用 DNS._resolvDNSByLocalServer
func (d DNS) resolvDNSByLocalServer(e *NetArg) (err error) {
	if dnsLocalCache.Exsit(e.Domain) {
		e.IP, err = dnsLocalCache.Wait4IP(e.Domain)
		if err != nil {
			return errors.New("resolvDNSByLocalServer -> " + err.Error())
		}
		return nil
	}
	dnsLocalCache.Add(e.Domain)
	err = d._resolvDNSByLocalServer(e)
	if err != nil {
		dnsLocalCache.Delete(e.Domain)
		return errors.New("resolvDNSByLocalServer -> " + err.Error())
	}
	dnsLocalCache.Update(e.Domain, e.IP)
	return nil
}

// _resolvDNSByLocalServer 本地解析DNS
// 实际完成DNS的解析动作
func (d DNS) _resolvDNSByLocalServer(e *NetArg) (err error) {
	e.IP, err = mynet.ResolvIPv4(e.Domain)
	if err != nil {
		return errors.New("_resolvDNSByLocalServer -> " + err.Error())
	}
	return nil
}

// IsWhiteDomain 判断域名是否是白名域名
func IsWhiteDomain(host string) (isWhite bool) {
	for _, line := range WhitelistDomains {
		if strings.HasSuffix(host, line) {
			return true
		}
	}
	return false
}
