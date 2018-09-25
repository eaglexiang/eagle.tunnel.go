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

// ET请求的类型
const (
	EtTCP = iota
	EtDNS
	EtLOCATION
	EtUNKNOWN
)

// 代理的状态
const (
	ProxyENABLE = iota
	ProxySMART
)

var protocolVersion, _ = CreateVersion("1.1")
var version, _ = CreateVersion("0.2")

// WhitelistDomains 需要被智能解析的DNS域名列表
var WhitelistDomains []string
var insideCache = sync.Map{}
var dnsRemoteCache = sync.Map{}
var hostsCache = make(map[string]string)

// EagleTunnel 遵循ET协议的数据隧道
type EagleTunnel struct {
}

func (et *EagleTunnel) handle(request Request, tunnel *Tunnel) (willContinue bool) {
	var result bool
	args := strings.Split(request.RequestMsgStr, " ")
	isVersionOk := et.checkVersionOfReq(args, tunnel)
	if isVersionOk {
		tunnel.encryptLeft = true
		isUserOk := checkUserOfReq(tunnel)
		if isUserOk {
			buffer := make([]byte, 1024)
			count, _ := tunnel.readLeft(buffer)
			if count > 0 {
				req := string(buffer[:count])
				args := strings.Split(req, " ")
				reqType := ParseEtType(args[0])
				switch reqType {
				case EtDNS:
					et.handleDNSReq(args, tunnel)
				case EtTCP:
					result = et.handleTCPReq(args, tunnel) == nil
				case EtLOCATION:
					et.handleLocationReq(args, tunnel)
				default:
				}
			}
		}
	}
	return result
}

// send 发送ET请求
func (et *EagleTunnel) send(e *NetArg) (succeed bool) {
	var result bool
	switch e.theType {
	case EtDNS:
		result = et.sendDNSReq(e)
	case EtTCP:
		result = et.sendTCPReq(e) == nil
	case EtLOCATION:
		result = et.sendLocationReq(e) == nil
	default:
	}
	return result
}

func (et *EagleTunnel) sendDNSReq(e *NetArg) (succeed bool) {
	var result bool
	ip, result := hostsCache[e.domain]
	if result {
		e.ip = ip
	} else {
		switch PROXY_STATUS {
		case ProxySMART:
			white := et.isWhiteDomain(e.domain)
			if white {
				result = et.resolvDNSByProxy(e) == nil
			} else {
				result = et.resolvDNSByLocal(e, true) == nil
			}
		case ProxyENABLE:
			result = et.resolvDNSByProxy(e) == nil
		default:
		}
	}
	return result
}

func (et *EagleTunnel) resolvDNSByProxy(e *NetArg) error {
	var err error
	_cache, ok := dnsRemoteCache.Load(e.domain)
	if ok { // found cache
		cache := _cache.(*DNSCache)
		if !cache.OverTTL() {
			e.ip = cache.ip
		} else { // cache's timed out
			err = et._resolvDNSByProxy(e)
		}
	} else { // not found
		err = et._resolvDNSByProxy(e)
	}
	return err
}

func (et *EagleTunnel) _resolvDNSByProxy(e *NetArg) error {
	tunnel := Tunnel{}
	err := et.connect2Relayer(&tunnel)
	if err != nil {
		return err
	}
	defer tunnel.close()
	req := FormatEtType(EtDNS) + " " + e.domain
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
	cache := CreateDNSCache(e.domain, e.ip)
	dnsRemoteCache.Store(cache.domain, cache)
	return err
}

func (et *EagleTunnel) resolvDNSByLocal(e *NetArg, recursive bool) error {
	err := _resolvDNSByLocal(e)

	if err != nil {
		err = et.resolvDNSByProxy(e)
	} else {
		if recursive {
			err1 := et.sendLocationReq(e)
			if err1 == nil {
				if !e.boolObj {
					ne := NetArg{domain: e.domain}
					err1 = et.resolvDNSByProxy(&ne)
					if err1 == nil {
						e.ip = ne.ip
					}
				}
			}
		}
	}
	return err
}

func _resolvDNSByLocal(e *NetArg) error {
	addrs, err := net.LookupHost(e.domain)
	if err == nil {
		var ok bool
		for _, addr := range addrs {
			ip := net.ParseIP(addr)
			if ip.To4() != nil {
				e.ip = addr
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
	err = checkVersionOfRelayer(tunnel)
	if err != nil {
		return err
	}
	tunnel.encryptRight = true
	err = checkUserOfLocal(tunnel)
	return err
}

func checkVersionOfRelayer(tunnel *Tunnel) error {
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

func checkUserOfLocal(tunnel *Tunnel) error {
	var err error
	if LocalUser == nil {
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

func checkUserOfReq(tunnel *Tunnel) (isValid bool) {
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

func (et *EagleTunnel) sendTCPReq(e *NetArg) error {
	var err error
	switch PROXY_STATUS {
	case ProxySMART:
		var inside bool
		err = et.sendLocationReq(e)
		if err == nil {
			inside = e.boolObj
		} else {
			inside = false
		}
		if inside {
			err = et.sendTCPReq2Server(e)
		} else {
			err = et.sendTCPReq2Remote(e)
		}
	case ProxyENABLE:
		err = et.sendTCPReq2Remote(e)
	}
	return err
}

func (et *EagleTunnel) sendTCPReq2Remote(e *NetArg) error {
	err := et.connect2Relayer(e.tunnel)
	if err != nil {
		return err
	}
	req := FormatEtType(EtTCP) + " " + e.ip + " " + strconv.Itoa(e.port)
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

// check
func (et *EagleTunnel) checkInsideByRemote(e *NetArg) error {
	tunnel := Tunnel{}
	err := et.connect2Relayer(&tunnel)
	if err != nil {
		return err
	}
	defer tunnel.close()
	req := FormatEtType(EtLOCATION) + " " + e.ip
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

func (et *EagleTunnel) sendTCPReq2Server(e *NetArg) error {
	ipe := e.ip + ":" + strconv.Itoa(e.port)
	conn, err := net.DialTimeout("tcp", ipe, 5*time.Second)
	if err != nil {
		return err
	}
	e.tunnel.right = &conn
	e.tunnel.encryptRight = false
	return err
}

// ParseEtType 得到字符串对应的ET请求类型
func ParseEtType(src string) int {
	var result int
	switch src {
	case "DNS":
		result = EtDNS
	case "TCP":
		result = EtTCP
	case "LOCATION":
		result = EtLOCATION
	default:
		result = EtUNKNOWN
	}
	return result
}

// FormatEtType 得到ET请求类型对应的字符串
func FormatEtType(src int) string {
	var result string
	switch src {
	case EtDNS:
		result = "DNS"
	case EtTCP:
		result = "TCP"
	case EtLOCATION:
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

func (et *EagleTunnel) handleDNSReq(reqs []string, tunnel *Tunnel) {
	if len(reqs) >= 2 {
		domain := reqs[1]
		e := NetArg{domain: domain}
		err := et.resolvDNSByLocal(&e, false)
		if err == nil {
			tunnel.writeLeft([]byte(e.ip))
		}
	}
}

func (et *EagleTunnel) handleTCPReq(reqs []string, tunnel *Tunnel) error {
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
	err = et.sendTCPReq2Server(&e)
	if err == nil {
		tunnel.writeLeft([]byte("ok"))
	} else {
		tunnel.writeLeft([]byte("nok"))
	}
	return err
}
