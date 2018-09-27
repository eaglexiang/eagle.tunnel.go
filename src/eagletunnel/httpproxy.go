package eagletunnel

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"../eaglelib"
)

// HTTP请求的类型
const (
	HTTPCONNECT = iota
	HTTPOTHERS
	HTTPERROR
)

// HTTPProxy HTTP代理
type HTTPProxy struct {
}

func (conn *HTTPProxy) handle(request Request, tunnel *eaglelib.Tunnel) bool {
	var result bool
	reqStr := request.RequestMsgStr
	reqType, host, port := dismantle(reqStr)
	_port, err := strconv.Atoi(port)
	if err == nil {
		if reqType != HTTPERROR && host != "" && _port > 0 {
			sender := EagleTunnel{}
			e := NetArg{domain: host, port: _port, tunnel: tunnel}
			ip := net.ParseIP(host)
			if ip == nil {
				e.TheType = EtDNS
				sender.Send(&e)
			} else {
				e.ip = e.domain
			}
			e.TheType = EtTCP
			ok := sender.Send(&e)
			if ok {
				if reqType == HTTPCONNECT {
					re443 := "HTTP/1.1 200 Connection Established\r\n\r\n"
					_, err = tunnel.WriteLeft([]byte(re443))
				} else {
					newReq := createNewRequest(reqStr)
					_, err = tunnel.WriteRight([]byte(newReq))
				}
				result = err == nil
			}
		}
	}
	return result
}

func dismantle(request string) (int, string, string) {
	var reqType int
	var host string
	var port string
	lines := strings.Split(request, "\r\n")
	args := strings.Split(lines[0], " ")
	if len(args) >= 3 {
		switch args[0] {
		case "CONNECT":
			reqType = HTTPCONNECT
		case "OPTIONS", "HEAD", "GET", "POST", "PUT", "DELETE", "TRACE":
			reqType = HTTPOTHERS
		default:
			reqType = HTTPERROR
		}
		u, err := url.Parse(args[1])
		if err == nil {
			if u.Host == "" {
				args[1] = "http://" + args[1]
				u, err = url.Parse(args[1])
			}
		}
		if err == nil {
			host = u.Host
			hostLines := strings.Split(host, ":")
			if len(hostLines) == 2 {
				host = hostLines[0]
				port = hostLines[1]
			}
			addr := net.ParseIP(host)
			if addr == nil {
				e := NetArg{domain: host, TheType: EtDNS}
				conn := EagleTunnel{}
				if conn.Send(&e) {
					host = e.ip
				}
			}
			if port == "" {
				port = u.Port()
			}
			if port == "" {
				switch u.Scheme {
				case "https":
					port = "443"
				case "http":
					port = "80"
				case "ftp":
					port = "21"
				default:
					port = "80"
				}
			}
		}
	}
	return reqType, host, port
}

func createNewRequest(oldRequest string) string {
	newReq := ""
	lines := strings.Split(oldRequest, "\r\n")
	firstLine := lines[0]
	argsOfFirstLine := strings.Split(firstLine, " ")
	if len(argsOfFirstLine) == 3 {
		u, err := url.Parse(argsOfFirstLine[1])
		if err == nil {
			path := u.Path
			if u.RawQuery != "" {
				path += "?" + u.RawQuery
			}
			argsOfFirstLine[1] = path
			newFirstLine := fmt.Sprintf("%s %s %s",
				argsOfFirstLine[0],
				argsOfFirstLine[1],
				argsOfFirstLine[2])
			newReq = newFirstLine
			for _, line := range lines[1:] {
				if strings.HasPrefix(line, "Proxy-Connection:") {
					connection := strings.TrimPrefix(line, "Proxy-Connection:")
					newReq += "\r\nConnection:" + connection
				} else {
					newReq += "\r\n" + line
				}
			}
		} else {
			fmt.Println(oldRequest)
		}
	}
	return newReq
}
