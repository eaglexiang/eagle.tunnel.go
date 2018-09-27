package eagletunnel

import (
	"strings"

	"../eaglelib"
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
		err := resolvDNSByLocal(&e, false)
		if err == nil {
			tunnel.WriteLeft([]byte(e.ip))
		}
	}
}

// Send 发送ET-DNS请求
func (ed *ETDNS) Send(e *NetArg) bool {
	var result bool
	ip, result := hostsCache[e.domain]
	if result {
		e.ip = ip
	} else {
		switch ProxyStatus {
		case ProxySMART:
			white := IsWhiteDomain(e.domain)
			if white {
				result = resolvDNSByProxy(e) == nil
			} else {
				result = resolvDNSByLocal(e, true) == nil
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
			e.ip = cache.IP
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
	e.ip = string(buffer[:count])
	cache := eaglelib.CreateDNSCache(e.domain, e.ip)
	dnsRemoteCache.Store(cache.Domain, cache)
	return err
}

func resolvDNSByLocal(e *NetArg, recursive bool) error {
	err := ResolvDNSByLocal(e)

	if err != nil {
		err = resolvDNSByProxy(e)
	} else {
		if recursive {
			el := ETLocation{}
			ok := el.Send(e)
			if ok {
				if !e.boolObj {
					ne := NetArg{domain: e.domain}
					err1 := resolvDNSByProxy(&ne)
					if err1 == nil {
						e.ip = ne.ip
					}
				}
			}
		}
	}
	return err
}
