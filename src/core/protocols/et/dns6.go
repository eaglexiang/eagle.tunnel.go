/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-13 18:54:13
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-03-17 17:06:46
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
	e := NetArg{Domain: reqs[1]}
	err := d.resolvDNS6ByLocalServer(&e)
	if err != nil {
		return err
	}
	tunnel.WriteLeft([]byte(e.IP))
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

	return d.proxySend(et, e)
}

// smartSend 智能模式
// 智能模式会先检查域名是否存在于白名单
// 白名单内域名将转入强制代理模式
func (d DNS6) smartSend(et *ET, e *NetArg) (err error) {
	if IsWhiteDomain(e.Domain) {
		err = d.resolvDNS6ByProxy(et, e)
	} else {
		err = d.resolvDNS6ByLocalClient(et, e)
	}
	return
}

// proxySend 强制代理模式
func (d DNS6) proxySend(et *ET, e *NetArg) error {
	return d.resolvDNS6ByProxy(et, e)
}

// Type ET子协议类型
func (d DNS6) Type() int {
	return EtDNS6
}

// resolvDNS6ByProxy 使用代理服务器进行DNS6的解析
// 此函数主要完成缓存功能
// 当缓存不命中则调用 DNS6._resolvDNSByProxy
func (d DNS6) resolvDNS6ByProxy(et *ET, e *NetArg) (err error) {
	node, loaded := dns6RemoteCache.Get(e.Domain)
	if loaded {
		e.IP, err = node.Wait()
	} else {
		err = d._resolvDNS6ByProxy(et, e)
		if err != nil {
			dns6RemoteCache.Delete(e.Domain)
		} else {
			dns6RemoteCache.Update(e.Domain, e.IP)
		}
	}
	return
}

// _resolvDNS6ByProxy 使用代理服务器进行DNS6的解析
// 实际完成DNS查询操作
func (d DNS6) _resolvDNS6ByProxy(et *ET, e *NetArg) error {
	req := FormatEtType(EtDNS6) + " " + e.Domain
	e.IP = sendQueryReq(et, req)
	ip := net.ParseIP(e.IP)
	if ip == nil {
		logger.Warning("resolv dns6 by proxy: ", e.Domain, " -> ",
			e.IP)
	}
	return nil
}

// resolvDNS6ByLocalClient 本地解析DNS6
// 此函数由客户端使用
// 此函数主要完成缓存功能
// 当缓存不命中则进一步调用 DNS6._resolvDNS6ByLocalClient
func (d DNS6) resolvDNS6ByLocalClient(et *ET, e *NetArg) (err error) {
	node, loaded := dns6LocalCache.Get(e.Domain)
	if loaded {
		e.IP, err = node.Wait()
	} else {
		err = d._resolvDNS6ByLocalClient(et, e)
		if err != nil {
			dns6LocalCache.Delete(e.Domain)
		} else {
			dns6LocalCache.Update(e.Domain, e.IP)
		}
	}
	return
}

// _resolvDNS6ByLocalClient 本地解析DNS6
// 实际完成DNS6的解析动作
func (d DNS6) _resolvDNS6ByLocalClient(et *ET, e *NetArg) (err error) {
	e.IP, err = mynet.ResolvIPv6(e.Domain)
	// 本地解析失败应该让用户察觉，手动添加DNS白名单
	if err != nil {
		logger.Warning("fail to resolv DNS6 by local: ", e.Domain,
			" , consider adding this domain to your whitelist_domain.txt")
		return err
	}

	// 判断IP所在位置是否适合代理
	l := et.subSenders[EtLOCATION].(Location)
	l.Send(et, e)
	if !l.CheckProxyByLocation(e.Location) {
		return nil
	}
	// 需要代理：更新IP为Relayer端的解析结果
	ne := NetArg{Domain: e.Domain}
	err = d.resolvDNS6ByProxy(et, &ne)
	e.IP = ne.IP
	return err
}

// resolvDNS6ByLocalServer 本地解析DNS6
// 此函数由服务端使用
// 此函数完成缓存相关的工作
// 当缓存不命中则进一步调用 DNS6._resolvDNSByLocalServer
func (d DNS6) resolvDNS6ByLocalServer(e *NetArg) (err error) {
	node, loaded := dns6LocalCache.Get(e.Domain)
	if loaded {
		e.IP, err = node.Wait()
	} else {
		err = d._resolvDNS6ByLocalServer(e)
		if err != nil {
			dns6LocalCache.Delete(e.Domain)
		} else {
			dns6LocalCache.Update(e.Domain, e.IP)
		}
	}
	return
}

// _resolvDNS6ByLocalServer 本地解析DNS6
// 实际完成DNS6的解析动作
func (d DNS6) _resolvDNS6ByLocalServer(e *NetArg) (err error) {
	e.IP, err = mynet.ResolvIPv6(e.Domain)
	return err
}
