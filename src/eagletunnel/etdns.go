package eagletunnel

import (
	"errors"
	"net"
	"strings"

	"github.com/eaglexiang/eagle.lib.go/src"
)

// ETDNS ET-DNS子协议的实现
type ETDNS struct {
}

// Handle 处理ET-DNS请求
func (ed *ETDNS) Handle(req Request, tunnel *eaglelib.Tunnel) {
	reqs := strings.Split(req.RequestMsgStr, " ")
	if len(reqs) >= 2 {
		domain := reqs[1]
		e := NetArg{domain: domain}
		err := resolvDNSByLocal(&e)
		if err == nil {
			tunnel.WriteLeft([]byte(e.IP))
		}
	}
}

// Send 发送ET-DNS请求
func (ed *ETDNS) Send(e *NetArg) bool {
	var result bool
	ip, result := hostsCache[e.domain]
	if result {
		e.IP = ip
	} else {
		switch ProxyStatus {
		case ProxySMART:
			white := IsWhiteDomain(e.domain)
			if white {
				result = resolvDNSByProxy(e) == nil
			} else {
				result = resolvDNSByLocal(e) == nil
			}
		case ProxyENABLE:
			result = resolvDNSByProxy(e) == nil
		default:
		}
	}
	return result
}

func resolvDNSByProxy(e *NetArg) error {
	var err error
	_cache, ok := dnsRemoteCache.Load(e.domain)
	if ok { // found cache
		cache := _cache.(*eaglelib.DNSCache)
		if !cache.OverTTL() {
			e.IP = cache.IP
		} else { // cache's timed out
			err = _resolvDNSByProxy(e)
		}
	} else { // not found
		err = _resolvDNSByProxy(e)
	}
	return err
}

func _resolvDNSByProxy(e *NetArg) error {
	tunnel := eaglelib.Tunnel{}
	defer tunnel.Close()
	err := connect2Relayer(&tunnel)
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
	cache := eaglelib.CreateDNSCache(e.domain, e.IP)
	dnsRemoteCache.Store(cache.Domain, cache)
	return err
}

// resolvDNSByLocal 本地解析DNS，优先返回IPv4结果
func resolvDNSByLocal(e *NetArg) error {
	err := ResolvIPv4ByLocal(e)
	if err != nil {
		err = ResolvIPv6ByLocal(e)
	}

	if err != nil {
		// 本地解析失败不得已使用远端解析
		err = resolvDNSByProxy(e)
		return err
	}

	// 判断本地解析结果的所在位置
	el := ETLocation{}
	ok := el.Send(e)
	if !ok {
		return nil
	}
	if !e.boolObj {
		ne := NetArg{domain: e.domain}
		err = resolvDNSByProxy(&ne)
		if err == nil {
			e.IP = ne.IP
		}
	}
	return nil
}
