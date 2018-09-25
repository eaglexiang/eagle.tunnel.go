package eagletunnel

import (
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
)

// WhitelistDomains 需要被智能解析的DNS域名列表
var WhitelistDomains []string

// ResolvDNSByLocal 本地解析DNS
func ResolvDNSByLocal(e *NetArg) error {
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

// CheckInsideByLocal 本地解析IP所在地
func CheckInsideByLocal(ip string) (bool, error) {
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

// IsWhiteDomain 判断域名是否是白名域名
func IsWhiteDomain(host string) (isWhite bool) {
	var white bool
	for _, line := range WhitelistDomains {
		if strings.HasSuffix(host, line) {
			white = true
			break
		}
	}
	return white
}
