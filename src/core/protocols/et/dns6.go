/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-13 18:54:13
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-02-21 19:02:05
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

var dns6RemoteCache = dnscache.CreateDNSCache()
var dns6LocalCache = dnscache.CreateDNSCache()

// DNS6 ET-DNS6子协议的实现
type DNS6 struct {
	arg *Arg
}

// Handle 处理ET-DNS6请求
func (d DNS6) Handle(req string, tunnel *mytunnel.Tunnel) error {
	reqs := strings.Split(req, " ")
	if len(reqs) < 2 {
		return errors.New("DNS6.Handle -> req is too short")
	}
	domain := reqs[1]
	e := NetArg{Domain: domain}
	err := d.resolvDNSByLocalServer(&e)
	if err != nil {
		return errors.New("DNS6.Handle -> " + err.Error())
	}
	_, err = tunnel.WriteLeft([]byte(e.IP))
	if err != nil {
		return errors.New("DNS6.Handle -> " + err.Error())
	}
	return nil
}

// Match 判断是否匹配
func (d DNS6) Match(req string) bool {
	args := strings.Split(req, " ")
	if args[0] == "DNS6" {
		return true
	}
	return false
}

// Send 发送ET-DNS6请求
func (d DNS6) Send(et *ET, e *NetArg) (err error) {
	ip, result := HostsCache[e.Domain]
	if result {
		e.IP = ip
		return nil
	}
	// 暂时全部使用Proxy模式，
	// 这是因为LOCATION指令暂不支持IPv6

	// switch d.ProxyStatus {
	// case ProxySMART:
	// 	err = d.smartSend(et, e)
	// 	if err != nil {
	// 		return errors.New("DNS6.Send -> " +
	// 			err.Error())
	// 	}
	// case ProxyENABLE:
	// 	err = d.proxySend(et, e)
	// 	if err != nil {
	// 		return errors.New("DNS6.Send -> " +
	// 			err.Error())
	// 	}
	// default:
	// 	err = errors.New("DNS6.Send -> invalid proxy-status")
	// }

	err = d.proxySend(et, e)

	if err != nil {
		return errors.New("DNS6.Send -> " +
			err.Error())
	}
	return nil
}

// smartSend 智能模式
// 智能模式会先检查域名是否存在于白名单
// 白名单内域名将转入强制代理模式
func (d DNS6) smartSend(et *ET, e *NetArg) error {
	white := IsWhiteDomain(e.Domain)
	if white {
		err := d.resolvDNSByProxy(et, e)
		if err != nil {
			return errors.New("DNS6.smartSend -> " + err.Error())
		}
		return nil
	}
	err := d.resolvDNSByLocalClient(et, e)
	if err != nil {
		return errors.New("DNS6.smartSend -> " + err.Error())
	}
	return nil
}

// proxySend 强制代理模式
func (d DNS6) proxySend(et *ET, e *NetArg) error {
	err := d.resolvDNSByProxy(et, e)
	if err != nil {
		return errors.New("DNS6.proxySend -> " + err.Error())
	}
	return nil
}

// Type ET子协议类型
func (d DNS6) Type() int {
	return EtDNS6
}

// resolvDNSByProxy 使用代理服务器进行DNS6的解析
// 此函数主要完成缓存功能
// 当缓存不命中则调用 DNS6._resolvDNSByProxy
func (d DNS6) resolvDNSByProxy(et *ET, e *NetArg) error {
	var err error
	if dns6RemoteCache.Exsit(e.Domain) {
		e.IP, err = dns6RemoteCache.Wait4IP(e.Domain)
		if err != nil {
			return errors.New("resolvDNSByProxy -> " + err.Error())
		}
		return nil
	}
	dns6RemoteCache.Add(e.Domain)
	err = d._resolvDNSByProxy(et, e)
	if err != nil {
		dns6RemoteCache.Delete(e.Domain)
		return errors.New("resolvDNSByProxy -> " + err.Error())
	}
	dns6RemoteCache.Update(e.Domain, e.IP)
	return nil
}

// _resolvDNSByProxy 使用代理服务器进行DNS6的解析
// 实际完成DNS查询操作
func (d DNS6) _resolvDNSByProxy(et *ET, e *NetArg) error {
	req := FormatEtType(EtDNS6) + " " + e.Domain
	reply := sendQueryReq(et, req)
	ip := net.ParseIP(reply)
	if ip == nil {
		return errors.New("_resolvDNSByProxy -> failed to resolv by remote: " +
			e.Domain + " -> " + reply)
	}
	e.IP = reply
	return nil
}

// resolvDNSByLocalClient 本地解析DNS6
// 此函数由客户端使用
// 此函数主要完成缓存功能
// 当缓存不命中则进一步调用 DNS6._resolvDNSByLocalClient
func (d DNS6) resolvDNSByLocalClient(et *ET, e *NetArg) (err error) {
	if dns6LocalCache.Exsit(e.Domain) {
		e.IP, err = dns6LocalCache.Wait4IP(e.Domain)
		if err != nil {
			return errors.New("resolvDNSByLocalClient -> " + err.Error())
		}
		return nil
	}
	dns6LocalCache.Add(e.Domain)
	err = d._resolvDNSByLocalClient(et, e)
	if err != nil {
		dns6LocalCache.Delete(e.Domain)
		return errors.New("resolvDNSByLocalClient -> " + err.Error())
	}
	dns6LocalCache.Update(e.Domain, e.IP)
	return nil
}

// _resolvDNSByLocalClient 本地解析DNS6
// 实际完成DNS6的解析动作
func (d DNS6) _resolvDNSByLocalClient(et *ET, e *NetArg) (err error) {
	e.IP, err = mynet.ResolvIPv6(e.Domain)
	// 本地解析失败应该让用户察觉，手动添加DNS白名单
	if err != nil {
		return errors.New("_resolvDNSByLocalClient -> fail to resolv DNS6 by local: " + e.Domain +
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

// resolvDNSByLocalServer 本地解析DNS6
// 此函数由服务端使用
// 此函数完成缓存相关的工作
// 当缓存不命中则进一步调用 DNS6._resolvDNSByLocalServer
func (d DNS6) resolvDNSByLocalServer(e *NetArg) (err error) {
	if dns6LocalCache.Exsit(e.Domain) {
		e.IP, err = dns6LocalCache.Wait4IP(e.Domain)
		if err != nil {
			return errors.New("resolvDNSByLocalServer -> " + err.Error())
		}
		return nil
	}
	dns6LocalCache.Add(e.Domain)
	err = d._resolvDNSByLocalServer(e)
	if err != nil {
		dns6LocalCache.Delete(e.Domain)
		return errors.New("resolvDNSByLocalServer -> " + err.Error())
	}
	dns6LocalCache.Update(e.Domain, e.IP)
	return nil
}

// _resolvDNSByLocalServer 本地解析DNS6
// 实际完成DNS6的解析动作
func (d DNS6) _resolvDNSByLocalServer(e *NetArg) (err error) {
	e.IP, err = mynet.ResolvIPv6(e.Domain)
	if err != nil {
		return errors.New("_resolvDNSByLocalServer -> " + err.Error())
	}
	return nil
}
