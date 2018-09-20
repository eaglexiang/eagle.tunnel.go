package eagletunnel

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	ET_TCP = iota
	ET_DNS
	ET_LOCATION
	ET_UNKNOWN
)

const (
	PROXY_ENABLE = iota
	PROXY_SMART
)

var protocolVersion, _ = CreateVersion("1.1")
var version, _ = CreateVersion("0.1")

var WhitelistDomains []string
var insideCache = sync.Map{}
var dnsRemoteCache = sync.Map{}
var dnsLocalCache = sync.Map{}
var hostsCache = make(map[string]string)

type EagleTunnel struct {
}

func (et *EagleTunnel) handle(request Request, tunnel *Tunnel) (willContinue bool) {
	var result bool
	args := strings.Split(request.RequestMsgStr, " ")
	isVersionOk := et.checkVersionOfReq(args, tunnel)
	if isVersionOk {
		tunnel.encryptLeft = true
		isUserOk := et.checkUserOfReq(tunnel)
		if isUserOk {
			buffer := make([]byte, 1024)
			count, _ := tunnel.readLeft(buffer)
			if count > 0 {
				req := string(buffer[:count])
				args := strings.Split(req, " ")
				reqType := ParseEtType(args[0])
				switch reqType {
				case ET_DNS:
					et.handleDnsReq(args, tunnel)
				case ET_TCP:
					result = et.handleTcpReq(args, tunnel) == nil
				case ET_LOCATION:
					et.handleLocationReq(args, tunnel)
				default:
				}
			}
		}
	}
	return result
}

func (conn *EagleTunnel) send(e *NetArg) (succeed bool) {
	var result bool
	switch e.theType {
	case ET_DNS:
		result = conn.sendDnsReq(e)
	case ET_TCP:
		result = conn.sendTcpReq(e) == nil
	case ET_LOCATION:
		result = conn.sendLocationReq(e) == nil
	default:
	}
	return result
}

func (et *EagleTunnel) sendDnsReq(e *NetArg) (succeed bool) {
	var result bool
	ip, result := hostsCache[e.domain]
	if result {
		e.ip = ip
	} else {
		switch PROXY_STATUS {
		case PROXY_SMART:
			white := et.isWhiteDomain(e.domain)
			if white {
				result = et.resolvDnsByProxy(e) == nil
			} else {
				result = et.resolvDnsByLocal(e, true) == nil
			}
		case PROXY_ENABLE:
			result = et.resolvDnsByProxy(e) == nil
		default:
		}
	}
	return result
}

func (et *EagleTunnel) resolvDnsByProxy(e *NetArg) error {
	var err error
	_cache, ok := dnsRemoteCache.Load(e.domain)
	if ok { // found cache
		cache := _cache.(*DnsCache)
		if cache.Check() { // cache is valid
			e.ip = cache.ip
		} else { // cache's timed out
			err = et._resolvDnsByProxy(e)
		}
	} else { // not found
		err = et._resolvDnsByProxy(e)
	}
	return err
}

func (et *EagleTunnel) _resolvDnsByProxy(e *NetArg) error {
	tunnel := Tunnel{}
	err := et.connect2Relayer(&tunnel)
	if err != nil {
		return err
	}
	defer tunnel.close()
	req := FormatEtType(ET_DNS) + " " + e.domain
	count, err := tunnel.writeRight([]byte(req))
	if err != nil {
		return err
	}
	buffer := make([]byte, 1024)
	count, err = tunnel.readRight(buffer)
	if err != nil {
		return err
	}
	e.ip = string(buffer[:count])
	cache := CreateDnsCache(e.domain, e.ip)
	dnsRemoteCache.Store(cache.domain, cache)
	return err
}

func (et *EagleTunnel) resolvDnsByLocal(e *NetArg, recursive bool) error {
	var err error
	_cache, ok := dnsLocalCache.Load(e.domain)
	if ok {
		cache := _cache.(*DnsCache)
		if cache.Check() {
			e.ip = cache.ip
		} else {
			err = et._resolvDnsByLocal(e)
		}
	} else {
		err = et._resolvDnsByLocal(e)
	}

	if err != nil {
		err = et.resolvDnsByProxy(e)
	} else {
		if recursive {
			err1 := et.sendLocationReq(e)
			if err1 == nil {
				if !e.boolObj {
					ne := NetArg{domain: e.domain}
					err1 = et.resolvDnsByProxy(&ne)
					if err1 == nil {
						e.ip = ne.ip
					}
				}
			}
		}
	}
	return err
}

func (et *EagleTunnel) _resolvDnsByLocal(e *NetArg) error {
	addrs, err := net.LookupHost(e.domain)
	if err == nil {
		var ok bool
		for _, addr := range addrs {
			ip := net.ParseIP(addr)
			if ip.To4() != nil {
				e.ip = addr
				cache := CreateDnsCache(e.domain, e.ip)
				dnsLocalCache.Store(cache.domain, cache)
				ok = true
				break
			}
		}
		if !ok {
			err = errors.New("ipv4 not found")
		}
	}
	return err
}

func (et *EagleTunnel) isWhiteDomain(host string) (isWhite bool) {
	var white bool
	for _, line := range WhitelistDomains {
		if strings.HasSuffix(host, line) {
			white = true
			break
		}
	}
	return white
}

func (et *EagleTunnel) connect2Relayer(tunnel *Tunnel) error {
	remoteIpe := RemoteAddr + ":" + RemotePort
	conn, err := net.DialTimeout("tcp", remoteIpe, 5*time.Second)
	if err != nil {
		return err
	}
	tunnel.right = &conn
	tunnel.encryptKey = EncryptKey
	err = et.checkVersionOfRelayer(tunnel)
	if err != nil {
		return err
	}
	tunnel.encryptRight = true
	err = et.checkUserOfLocal(tunnel)
	return err
}

func (et *EagleTunnel) checkVersionOfRelayer(tunnel *Tunnel) error {
	req := "eagle_tunnel " + protocolVersion.raw + " simple"
	count, err := tunnel.writeRight([]byte(req))
	if err != nil {
		return err
	}
	var buffer = make([]byte, 1024)
	count, err = tunnel.readRight(buffer)
	if err != nil {
		return err
	}
	reply := string(buffer[0:count])
	if reply != "valid valid valid" {
		return errors.New(reply)
	}
	return err
}

func (et *EagleTunnel) checkVersionOfReq(headers []string, tunnel *Tunnel) (isValid bool) {
	var result bool
	if len(headers) >= 3 {
		replys := make([]string, 3)
		if headers[0] == "eagle_tunnel" {
			replys[0] = "valid"
		} else {
			replys[0] = "invalid"
		}
		versionOfReq, err := CreateVersion(headers[1])
		if err == nil {
			if protocolVersion.isBThanOrE2(&versionOfReq) {
				replys[1] = "valid"
			} else {
				replys[1] = "server is elder"
			}
		} else {
			replys[1] = err.Error()
		}
		if headers[2] == "simple" {
			replys[2] = "valid"
		} else {
			replys[2] = "invalid"
		}
		reply := replys[0] + " " + replys[1] + " " + replys[2]
		count, _ := tunnel.writeLeft([]byte(reply))
		result = count == 17 // length of "valid valid valid"
	}
	return result
}

func (et *EagleTunnel) checkUserOfLocal(tunnel *Tunnel) error {
	var err error
	if LocalUser.ID == "" {
		return nil // no need to check
	}
	user := LocalUser.toString()
	var count int
	count, err = tunnel.writeRight([]byte(user))
	if err != nil {
		return err
	}
	buffer := make([]byte, 1024)
	count, err = tunnel.readRight(buffer)
	if err != nil {
		return err
	}
	reply := string(buffer[:count])
	if reply != "valid" {
		err = errors.New(reply)
	} else {
		LocalUser.addTunnel(tunnel)
	}
	return err
}

func (et *EagleTunnel) checkUserOfReq(tunnel *Tunnel) (isValid bool) {
	var result bool
	if EnableUserCheck {
		buffer := make([]byte, 1024)
		count, _ := tunnel.readLeft(buffer)
		if count > 0 {
			userStr := string(buffer[:count])
			user, err := ParseEagleUser(userStr, (*tunnel.left).RemoteAddr())
			if err == nil {
				err = Users[user.ID].CheckAuth(user)
			}
			if err == nil {
				reply := "valid"
				count, _ = tunnel.writeLeft([]byte(reply))
				result = count == 5
				if result {
					Users[user.ID].addTunnel(tunnel)
				}
			} else {
				reply := err.Error()
				_, _ = tunnel.writeLeft([]byte(reply))
			}
		}
	} else {
		result = true
	}
	return result
}

func (et *EagleTunnel) sendTcpReq(e *NetArg) error {
	var err error
	switch PROXY_STATUS {
	case PROXY_SMART:
		var inside bool
		err = et.sendLocationReq(e)
		if err == nil {
			inside = e.boolObj
		} else {
			inside = false
		}
		if inside {
			err = et.sendTcpReq2Server(e)
		} else {
			err = et.sendTcpReq2Remote(e)
		}
	case PROXY_ENABLE:
		err = et.sendTcpReq2Remote(e)
	}
	return err
}

func (et *EagleTunnel) sendTcpReq2Remote(e *NetArg) error {
	err := et.connect2Relayer(e.tunnel)
	if err != nil {
		return err
	}
	req := FormatEtType(ET_TCP) + " " + e.ip + " " + strconv.Itoa(e.port)
	count, err := e.tunnel.writeRight([]byte(req))
	if err != nil {
		return err
	}
	buffer := make([]byte, 1024)
	count, err = e.tunnel.readRight(buffer)
	if err != nil {
		return err
	}
	reply := string(buffer[:count])
	if reply != "ok" {
		err = errors.New("failed 2 connect 2 server by relayer")
	}
	return err
}

func (et *EagleTunnel) sendLocationReq(e *NetArg) error {
	var err error
	_inside, ok := insideCache.Load(e.ip)
	if ok {
		e.boolObj, _ = _inside.(bool)
	} else {
		err = et.checkInsideByRemote(e)
		if err == nil {
			insideCache.Store(e.ip, e.boolObj)
		} else {
			var inside bool
			inside, err = et.checkInsideByLocal(e.ip)
			if err == nil {
				e.boolObj = inside
				insideCache.Store(e.ip, e.boolObj)
			}
		}
	}
	return err
}

func (conn *EagleTunnel) checkInsideByRemote(e *NetArg) error {
	tunnel := Tunnel{}
	err := conn.connect2Relayer(&tunnel)
	if err != nil {
		return err
	}
	defer tunnel.close()
	req := FormatEtType(ET_LOCATION) + " " + e.ip
	var count int
	count, err = tunnel.writeRight([]byte(req))
	if err != nil {
		return err
	}
	buffer := make([]byte, 1024)
	count, err = tunnel.readRight(buffer)
	if err != nil {
		return err
	}
	e.boolObj, err = strconv.ParseBool(string(buffer[0:count]))
	return err
}

func (et *EagleTunnel) sendTcpReq2Server(e *NetArg) error {
	ipe := e.ip + ":" + strconv.Itoa(e.port)
	conn, err := net.DialTimeout("tcp", ipe, 5*time.Second)
	if err != nil {
		return err
	}
	e.tunnel.right = &conn
	e.tunnel.encryptRight = false
	return err
}

func ParseEtType(src string) int {
	var result int
	switch src {
	case "DNS":
		result = ET_DNS
	case "TCP":
		result = ET_TCP
	case "LOCATION":
		result = ET_LOCATION
	default:
		result = ET_UNKNOWN
	}
	return result
}

func FormatEtType(src int) string {
	var result string
	switch src {
	case ET_DNS:
		result = "DNS"
	case ET_TCP:
		result = "TCP"
	case ET_LOCATION:
		result = "LOCATION"
	default:
	}
	return result
}

func (et *EagleTunnel) handleLocationReq(reqs []string, tunnel *Tunnel) {
	if len(reqs) >= 2 {
		var reply string
		ip := reqs[1]
		_inside, ok := insideCache.Load(ip)
		if ok {
			inside := _inside.(bool)
			reply = strconv.FormatBool(inside)
		} else {
			var err error
			inside, err := et.checkInsideByLocal(ip)
			if err != nil {
				reply = fmt.Sprint(err)
			} else {
				reply = strconv.FormatBool(inside)
				insideCache.Store(ip, inside)
			}
		}
		tunnel.writeLeft([]byte(reply))
	}
}

func (et *EagleTunnel) checkInsideByLocal(ip string) (bool, error) {
	var inside bool
	req := "https://ip2c.org/" + ip
	res, err := http.Get(req)
	if err != nil {
		return inside, err
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	bodyStr := string(body)
	if err == nil {
		switch bodyStr {
		case "0;;;WRONG INPUT":
			err = errors.New("0;;;WRONG INPUT")
		case "1;ZZ;ZZZ;Reserved", "1;CN;CHN;China":
			inside = true
		default:
		}
	}
	return inside, err
}

func (et *EagleTunnel) handleDnsReq(reqs []string, tunnel *Tunnel) {
	if len(reqs) >= 2 {
		domain := reqs[1]
		e := NetArg{domain: domain}
		err := et.resolvDnsByLocal(&e, false)
		if err == nil {
			tunnel.writeLeft([]byte(e.ip))
		}
	}
}

func (et *EagleTunnel) handleTcpReq(reqs []string, tunnel *Tunnel) error {
	if len(reqs) < 3 {
		return errors.New("invalid reqs")
	}
	ip := reqs[1]
	_port := reqs[2]
	port, err := strconv.ParseInt(_port, 10, 32)
	if err != nil {
		return err
	}
	e := NetArg{ip: ip, port: int(port), tunnel: tunnel}
	err = et.sendTcpReq2Server(&e)
	if err == nil {
		tunnel.writeLeft([]byte("ok"))
	} else {
		tunnel.writeLeft([]byte("nok"))
	}
	return err
}
