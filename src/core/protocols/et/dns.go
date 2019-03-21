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

	logger "github.com/eaglexiang/eagle.tunnel.go/src/logger"
	"github.com/eaglexiang/go-textcache"
	cache "github.com/eaglexiang/go-textcache"
	mytunnel "github.com/eaglexiang/go-tunnel"
)

// HostsCache 本地Hosts
var HostsCache = make(map[string]string)

// WhitelistDomains 需要被智能解析的DNS域名列表
var WhitelistDomains []string

// dNS ET-DNS子协议的实现
type dNS struct {
	dnsType        int
	arg            *Arg
	dnsRemoteCache *cache.TextCache
	dnsLocalCache  *cache.TextCache
	dnsResolver    func(string) (string, error)
}

// Handle 处理ET-DNS请求
func (d dNS) Handle(req string, tunnel *mytunnel.Tunnel) error {
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

// Send 发送ET-DNS请求
func (d dNS) Send(et *ET, e *NetArg) (err error) {
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
func (d dNS) smartSend(et *ET, e *NetArg) (err error) {
	white := IsWhiteDomain(e.Domain)
	if white {
		err = d.resolvDNSByProxy(et, e)
	} else {
		err = d.resolvDNSByLocal(e)
		// 判断IP所在位置是否适合代理
		l := et.subSenders[EtLOCATION].(location)
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
func (d dNS) proxySend(et *ET, e *NetArg) error {
	return d.resolvDNSByProxy(et, e)
}

// Type ET子协议类型
func (d dNS) Type() int {
	return d.dnsType
}

// Name ET子协议的名字
func (d dNS) Name() string {
	return FormatEtType(d.dnsType)
}

func (d *dNS) getCacheNodeOfRemote(domain string) (node *textcache.CacheNode, loaded bool) {
	if d.dnsRemoteCache == nil {
		d.dnsRemoteCache = cache.CreateTextCache()
	}
	return d.dnsRemoteCache.Get(domain)
}

func (d *dNS) getCacheNodeOfLocal(domain string) (node *cache.CacheNode, loaded bool) {
	if d.dnsLocalCache == nil {
		d.dnsLocalCache = cache.CreateTextCache()
	}
	return d.dnsLocalCache.Get(domain)
}

// resolvDNSByProxy 使用代理服务器进行DNS的解析
// 此函数主要完成缓存功能
// 当缓存不命中则调用 dNS._resolvDNSByProxy
func (d dNS) resolvDNSByProxy(et *ET, e *NetArg) (err error) {
	node, loaded := d.getCacheNodeOfRemote(e.Domain)
	if loaded {
		e.IP, err = node.Wait()
	} else {
		err = d._resolvDNSByProxy(et, e)
		if err != nil {
			d.dnsRemoteCache.Delete(e.Domain)
		} else {
			node.Update(e.IP)
		}
	}
	return err
}

// _resolvDNSByProxy 使用代理服务器进行DNS的解析
// 实际完成DNS查询操作
func (d dNS) _resolvDNSByProxy(et *ET, e *NetArg) (err error) {
	req := FormatEtType(EtDNS) + " " + e.Domain
	e.IP, err = sendQueryReq(et, req)
	ip := net.ParseIP(e.IP)
	if ip == nil {
		logger.Warning("fail to resolv dns by proxy: ", e.Domain, " -> ", e.IP)
		return errors.New("invalid reply")
	}
	return nil
}

// resolvDNSByLocal 本地解析DNS
// 此函数主要完成缓存功能
// 当缓存不命中则进一步调用 dNS._resolvDNSByLocalClient
func (d dNS) resolvDNSByLocal(e *NetArg) (err error) {
	node, loaded := d.getCacheNodeOfLocal(e.Domain)
	if loaded {
		e.IP, err = node.Wait()
	} else {
		err = d._resolvDNSByLocal(e)
		if err != nil {
			d.dnsLocalCache.Delete(e.Domain)
		} else {
			node.Update(e.IP)
		}
	}
	return err
}

// _resolvDNSByLocalClient 本地解析DNS
// 实际完成DNS的解析动作
func (d dNS) _resolvDNSByLocal(e *NetArg) (err error) {
	e.IP, err = d.dnsResolver(e.Domain)
	// 本地解析失败应该让用户察觉，手动添加DNS白名单
	if err != nil {
		logger.Warning("fail to resolv dns by local, ",
			"consider adding this domain to your whitelist_domain.txt: ",
			e.Domain)
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
