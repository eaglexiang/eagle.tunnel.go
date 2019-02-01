/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-13 18:54:13
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-22 19:33:39
 */

package et

import (
	"errors"
	"net"
	"strings"

	"github.com/eaglexiang/go-bytebuffer"

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
// 必须使用createDNS进行构造
type DNS struct {
	proxyStatus      int
	firstIPResolver  func(domain string) (ip string, err error)
	secondIPResolver func(domain string) (ip string, err error)
}

// createDNS 构造DNS
func createDNS(
	proxyStatus int,
	firstIPResolver string,
) DNS {
	dns := DNS{proxyStatus: proxyStatus}
	switch firstIPResolver {
	case "4":
		dns.firstIPResolver = mynet.ResolvIPv4
	case "6":
		dns.firstIPResolver = mynet.ResolvIPv6
	case "46":
		dns.firstIPResolver = mynet.ResolvIPv4
		dns.secondIPResolver = mynet.ResolvIPv6
	case "64":
		dns.firstIPResolver = mynet.ResolvIPv6
		dns.secondIPResolver = mynet.ResolvIPv4
	}
	return dns
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
	switch d.proxyStatus {
	case ProxySMART:
		err = d.smartSend(et, e)
	case ProxyENABLE:
		err = d.proxySend(et, e)
	default:
		err = errors.New("invalid proxy-status: " +
			FormatProxyStatus(d.proxyStatus))
	}

	if err != nil {
		return errors.New("DNS.Send -> " +
			err.Error())
	}
	return nil
}

// smartSend 智能模式
func (d *DNS) smartSend(et *ET, e *NetArg) error {
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
func (d *DNS) proxySend(et *ET, e *NetArg) error {
	err := d.resolvDNSByProxy(et, e)
	if err != nil {
		return errors.New("DNS.proxySend -> " + err.Error())
	}
	return nil
}

// ETType ET子协议类型
func (d DNS) ETType() int {
	return EtDNS
}

func (d *DNS) resolvDNSByProxy(et *ET, e *NetArg) error {
	var err error
	if dnsRemoteCache.Exsit(e.Domain) {
		e.IP, err = dnsRemoteCache.Wait4IP(e.Domain)
		if err != nil {
			return errors.New("resolvDNSByProxy -> " + err.Error())
		}
		return nil
	}
	dnsRemoteCache.Add(e.Domain)
	err = _resolvDNSByProxy(et, e)
	if err != nil {
		dnsRemoteCache.Delete(e.Domain)
		return errors.New("resolvDNSByProxy -> " + err.Error())
	}
	dnsRemoteCache.Update(e.Domain, e.IP)
	return nil
}

func _resolvDNSByProxy(et *ET, e *NetArg) error {
	// connect 2 relayer
	tunnel := mytunnel.GetTunnel()
	defer mytunnel.PutTunnel(tunnel)
	err := et.connect2Relayer(tunnel)
	if err != nil {
		return errors.New("_resolvDNSByProxy -> " + err.Error())
	}
	// send req
	req := FormatEtType(EtDNS) + " " + e.Domain
	_, err = tunnel.WriteRight([]byte(req))
	if err != nil {
		return errors.New("_resolvDNSByProxy -> " + err.Error())
	}
	// get reply
	buffer := bytebuffer.GetKBBuffer()
	defer bytebuffer.PutKBBuffer(buffer)
	buffer.Length, err = tunnel.ReadRight(buffer.Buf())
	if err != nil {
		return errors.New("_resolvDNSByProxy -> " + err.Error())
	}
	reply := buffer.String()
	ip := net.ParseIP(reply)
	if ip == nil {
		return errors.New("_resolvDNSByProxy -> failed to resolv by remote: " +
			e.Domain + " -> " + reply)
	}
	e.IP = reply
	return nil
}

// resolvDNSByLocalClient 本地解析DNS，如无ip-type配置，优先返回IPv4
func (d *DNS) resolvDNSByLocalClient(et *ET, e *NetArg) (err error) {
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

func (d *DNS) _resolvDNSByLocalClient(et *ET, e *NetArg) (err error) {
	if d.firstIPResolver == nil {
		return errors.New("ETDNS._resolvDNSByLocalClient -> " +
			"ETDNS.FirstIPResolver is nil")
	}
	e.IP, err = d.firstIPResolver(e.Domain)
	if err != nil {
		if d.secondIPResolver != nil {
			e.IP, err = d.secondIPResolver(e.Domain)
		}
	}

	// 本地解析失败应该让用户察觉，手动添加DNS白名单
	if err != nil {
		return errors.New("_resolvDNSByLocalClient -> fail to resolv DNS by local: " + e.Domain)
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

// resolvDNSByLocalServer 本地解析DNS，如无ip-type配置，优先返回IPv4
func (d *DNS) resolvDNSByLocalServer(e *NetArg) (err error) {
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

func (d *DNS) _resolvDNSByLocalServer(e *NetArg) (err error) {
	if d.firstIPResolver == nil {
		return errors.New("_resolvDNSByLocalServer -> " +
			"ETDNS.FirstIPResolver is nil")
	}
	e.IP, err = d.firstIPResolver(e.Domain)
	if err != nil {
		if d.secondIPResolver != nil {
			e.IP, err = d.secondIPResolver(e.Domain)
		}
	}

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
