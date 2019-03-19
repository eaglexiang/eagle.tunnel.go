/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-13 18:54:13
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-03-17 17:00:52
 */

package et

import (
	"errors"
	"net"
	"strings"

	"github.com/eaglexiang/eagle.tunnel.go/src/logger"

	dnscache "github.com/eaglexiang/go-dnscache"
	mytunnel "github.com/eaglexiang/go-tunnel"
)

// HostsCache 本地Hosts
var HostsCache = make(map[string]string)

// WhitelistDomains 需要被智能解析的DNS域名列表
var WhitelistDomains []string

// DNS ET-DNS子协议的实现
type DNS struct {
	arg            *Arg
	dnsRemoteCache *dnscache.DNSCache
	dnsLocalCache  *dnscache.DNSCache
	dnsResolver    func(string) (string, error)
}

// Handle 处理ET-DNS请求
func (d DNS) Handle(req string, tunnel *mytunnel.Tunnel) error {
	reqs := strings.Split(req, " ")
	if len(reqs) < 2 {
		return errors.New("ETDNS.Handle -> req is too short")
	}
	e := NetArg{NetConnArg: NetConnArg{Domain: reqs[1]}}
	err := d.resolvDNSByLocal(&e)
	if err != nil {
		return err
	}
	_, err = tunnel.WriteLeft([]byte(e.IP))
	if err != nil {
		return err
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
	switch d.arg.ProxyStatus {
	case ProxySMART:
		err = d.smartSend(et, e)
	case ProxyENABLE:
		err = d.proxySend(et, e)
	default:
		logger.Error("dns.send invalid proxy-status")
		err = errors.New("invalid proxy-status")
	}
	return err
}

// smartSend 智能模式
// 智能模式会先检查域名是否存在于白名单
// 白名单内域名将转入强制代理模式
func (d DNS) smartSend(et *ET, e *NetArg) (err error) {
	white := IsWhiteDomain(e.Domain)
	if white {
		err = d.resolvDNSByProxy(et, e)
	} else {
		err = d.resolvDNSByLocal(e)
		// 判断IP所在位置是否适合代理
		l := et.subSenders[EtLOCATION].(Location)
		l.Send(et, e)
		if !l.CheckProxyByLocation(e.Location) {
			return nil
		}
		// 更新IP为Relay端的解析结果
		ne := NetArg{NetConnArg: NetConnArg{Domain: e.Domain}}
		err = d.resolvDNSByProxy(et, &ne)
		e.IP = ne.IP
	}
	return err
}

// proxySend 强制代理模式
func (d DNS) proxySend(et *ET, e *NetArg) error {
	return d.resolvDNSByProxy(et, e)
}

// Type ET子协议类型
func (d DNS) Type() int {
	return EtDNS
}

func (d *DNS) getCacheNodeOfRemote(domain string) (node *dnscache.CacheNode, loaded bool) {
	if d.dnsRemoteCache == nil {
		d.dnsRemoteCache = dnscache.CreateDNSCache()
	}
	return d.dnsRemoteCache.Get(domain)
}

func (d *DNS) getCacheNodeOfLocal(domain string) (node *dnscache.CacheNode, loaded bool) {
	if d.dnsLocalCache == nil {
		d.dnsLocalCache = dnscache.CreateDNSCache()
	}
	return d.dnsLocalCache.Get(domain)
}

// resolvDNSByProxy 使用代理服务器进行DNS的解析
// 此函数主要完成缓存功能
// 当缓存不命中则调用 DNS._resolvDNSByProxy
func (d DNS) resolvDNSByProxy(et *ET, e *NetArg) (err error) {
	node, loaded := d.getCacheNodeOfRemote(e.Domain)
	if loaded {
		e.IP, err = node.Wait()
	} else {
		err = d._resolvDNSByProxy(et, e)
		if err != nil {
			d.dnsRemoteCache.Delete(e.Domain)
		} else {
			d.dnsRemoteCache.Update(e.Domain, e.IP)
		}
	}
	return err
}

// _resolvDNSByProxy 使用代理服务器进行DNS的解析
// 实际完成DNS查询操作
func (d DNS) _resolvDNSByProxy(et *ET, e *NetArg) error {
	req := FormatEtType(EtDNS) + " " + e.Domain
	e.IP = sendQueryReq(et, req)
	ip := net.ParseIP(e.IP)
	if ip == nil {
		logger.Warning("resolv dns by proxy: ", e.IP)
		return errors.New("invalid reply")
	}
	return nil
}

// resolvDNSByLocal 本地解析DNS
// 此函数主要完成缓存功能
// 当缓存不命中则进一步调用 DNS._resolvDNSByLocalClient
func (d DNS) resolvDNSByLocal(e *NetArg) (err error) {
	node, loaded := d.getCacheNodeOfLocal(e.Domain)
	if loaded {
		e.IP, err = node.Wait()
	} else {
		err = d._resolvDNSByLocal(e)
		if err != nil {
			d.dnsLocalCache.Delete(e.Domain)
		} else {
			d.dnsLocalCache.Update(e.Domain, e.IP)
		}
	}
	return err
}

// _resolvDNSByLocalClient 本地解析DNS
// 实际完成DNS的解析动作
func (d DNS) _resolvDNSByLocal(e *NetArg) (err error) {
	e.IP, err = d.dnsResolver(e.Domain)
	// 本地解析失败应该让用户察觉，手动添加DNS白名单
	if err != nil {
		logger.Warning("fail to resolv dns by local, ",
			"consider adding this domain to your whitelist_domain.txt: ",
			e.Domain)
		return err
	}
	return err
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
