package eagletunnel

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

const (
	HTTP_CONNECT = iota
	HTTP_OTHERS
	HTTP_ERROR
)

type HttpProxy struct {
}

func (conn *HttpProxy) handle(request Request, tunnel *Tunnel) bool {
	var result bool
	reqStr := request.RequestMsgStr
	reqType, host, port := conn.dismantle(reqStr)
	_port, err := strconv.Atoi(port)
	if err == nil {
		if reqType != HTTP_ERROR && host != "" && _port > 0 {
			sender := EagleTunnel{}
			e := NetArg{domain: host, port: _port, tunnel: tunnel}
			ip := net.ParseIP(host)
			if ip == nil {
				e.theType = EtDNS
				sender.send(&e)
			} else {
				e.ip = e.domain
			}
			e.theType = EtTCP
			ok := sender.send(&e)
			if ok {
				if reqType == HTTP_CONNECT {
					re443 := "HTTP/1.1 200 Connection Established\r\n\r\n"
					_, err = tunnel.writeLeft([]byte(re443))
				} else {
					newReq := conn.createNewRequest(reqStr)
					_, err = tunnel.writeRight([]byte(newReq))
				}
				result = err == nil
			}
		}
	}
	return result
}

func (conn *HttpProxy) dismantle(request string) (int, string, string) {
	var reqType int
	var host string
	var port string
	lines := strings.Split(request, "\r\n")
	args := strings.Split(lines[0], " ")
	if len(args) >= 3 {
		switch args[0] {
		case "CONNECT":
			reqType = HTTP_CONNECT
		case "OPTIONS", "HEAD", "GET", "POST", "PUT", "DELETE", "TRACE":
			reqType = HTTP_OTHERS
		default:
			reqType = HTTP_ERROR
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
				e := NetArg{domain: host, theType: EtDNS}
				conn := EagleTunnel{}
				if conn.send(&e) {
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

func (conn *HttpProxy) createNewRequest(oldRequest string) string {
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

func (conn *HttpProxy) getKeyValues(lines []string) (map[string]string, []string) {
	keyValues := make(map[string]string)
	keys := make([]string, 0)
	for _, line := range lines {
		keyValue := strings.Split(line, ": ")
		if len(keyValue) >= 2 {
			value := keyValue[1]
			for _, item := range keyValue[2:] {
				value += ": " + item
			}
			keys = append(keys, keyValue[0])
			keyValues[keyValue[0]] = value
		}
	}
	return keyValues, keys
}

func (conn *HttpProxy) exportKeyValues(keyValues *map[string]string, keys []string) string {
	var result string
	for _, key := range keys {
		result += key + ": " + (*keyValues)[key] + "\r\n"
	}
	return result
}
