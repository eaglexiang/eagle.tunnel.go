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
	arg *Arg
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
		err = d.resolvDNSByLocalClient(et, e)
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

// resolvDNSByProxy 使用代理服务器进行DNS的解析
// 此函数主要完成缓存功能
// 当缓存不命中则调用 DNS._resolvDNSByProxy
func (d DNS) resolvDNSByProxy(et *ET, e *NetArg) (err error) {
	node, loaded := dnsRemoteCache.Get(e.Domain)
	if loaded {
		e.IP, err = node.Wait()
	} else {
		err = d._resolvDNSByProxy(et, e)
		if err != nil {
			dnsRemoteCache.Delete(e.Domain)
		} else {
			dnsRemoteCache.Update(e.Domain, e.IP)
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

// resolvDNSByLocalClient 本地解析DNS
// 此函数由客户端使用
// 此函数主要完成缓存功能
// 当缓存不命中则进一步调用 DNS._resolvDNSByLocalClient
func (d DNS) resolvDNSByLocalClient(et *ET, e *NetArg) (err error) {
	node, loaded := dnsLocalCache.Get(e.Domain)
	if loaded {
		e.IP, err = node.Wait()
	} else {
		err = d._resolvDNSByLocalClient(et, e)
		if err != nil {
			dnsLocalCache.Delete(e.Domain)
		} else {
			dnsLocalCache.Update(e.Domain, e.IP)
		}
	}
	return err
}

// _resolvDNSByLocalClient 本地解析DNS
// 实际完成DNS的解析动作
func (d DNS) _resolvDNSByLocalClient(et *ET, e *NetArg) (err error) {
	e.IP, err = mynet.ResolvIPv4(e.Domain)
	// 本地解析失败应该让用户察觉，手动添加DNS白名单
	if err != nil {
		logger.Warning("fail to resolv dns by local, ",
			"consider adding this domain to your whitelist_domain.txt: ",
			e.Domain)
		return err
	}

	// 判断IP所在位置是否适合代理
	l := et.subSenders[EtLOCATION].(Location)
	l.Send(et, e)
	if !l.CheckProxyByLocation(e.Location) {
		return nil
	}
	// 更新IP为Relayer端的解析结果
	ne := NetArg{Domain: e.Domain}
	err = d.resolvDNSByProxy(et, &ne)
	e.IP = ne.IP
	return err
}

// resolvDNSByLocalServer 本地解析DNS
// 此函数由服务端使用
// 此函数完成缓存相关的工作
// 当缓存不命中则进一步调用 DNS._resolvDNSByLocalServer
func (d DNS) resolvDNSByLocalServer(e *NetArg) (err error) {
	node, loaded := dnsLocalCache.Get(e.Domain)
	if loaded {
		e.IP, err = node.Wait()
	} else {
		err = d._resolvDNSByLocalServer(e)
		if err != nil {
			dnsLocalCache.Delete(e.Domain)
		} else {
			dnsLocalCache.Update(e.Domain, e.IP)
		}
	}
	return
}

// _resolvDNSByLocalServer 本地解析DNS
// 实际完成DNS的解析动作
func (d DNS) _resolvDNSByLocalServer(e *NetArg) (err error) {
	e.IP, err = mynet.ResolvIPv4(e.Domain)
	if err != nil {
		logger.Warning(err)
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
