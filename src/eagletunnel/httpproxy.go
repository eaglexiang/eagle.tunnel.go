/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-04 14:30:39
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-04 18:40:37
 */

package eagletunnel

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"../eaglelib/src"
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

// Handle 处理HTTPProxy请求
func (conn *HTTPProxy) Handle(request Request, tunnel *eaglelib.Tunnel) error {
	ipOfReq := strings.Split((*tunnel.Left).RemoteAddr().String(), ":")[0]
	if !CheckPrivateIPv4(ipOfReq) {
		// 不接受来自公网IP的HTTP代理请求
		return errors.New("invalid source ip type: public " + ipOfReq)
	}

	reqStr := request.RequestMsgStr
	reqType, host, _port := dismantle(reqStr)
	port, err := strconv.Atoi(_port)
	if err != nil {
		return err
	}
	if host == "" || port <= 0 {
		return errors.New("invalid des: " + host + ":" + _port)
	}
	sender := EagleTunnel{}
	e := NetArg{domain: host, Port: port, tunnel: tunnel}
	// get ip
	ip := net.ParseIP(host)
	if ip == nil {
		e.TheType = EtDNS
		err = sender.Send(&e)
		if err != nil {
			return err
		}
	} else {
		e.IP = e.domain
	}
	// dail tcp
	switch reqType {
	case HTTPCONNECT:
		e.TheType = EtTCP
		err = sender.Send(&e)
		if err != nil {
			return err
		}
		re443 := "HTTP/1.1 200 Connection Established\r\n\r\n"
		_, err = tunnel.WriteLeft([]byte(re443))
		return err
	case HTTPOTHERS:
		e.TheType = EtTCP
		err = sender.Send(&e)
		if err != nil {
			return err
		}
		newReq := createNewRequest(reqStr)
		_, err = tunnel.WriteRight([]byte(newReq))
		return err
	default:
		return errors.New("invalid HTTP type: " + reqStr)
	}
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
				switch reqType {
				case HTTPCONNECT:
					args[1] = "https://" + args[1]
				case HTTPOTHERS:
					args[1] = "http://" + args[1]
				}
				u, err = url.Parse(args[1])
			}
		} else {
			args := strings.Split(args[1], ":")
			if len(args) == 2 {
				host = args[0]
				port = args[1]
			}
		}
		if err == nil {
			host = u.Host
			hostLines := strings.Split(host, ":")
			host = hostLines[0]
			if len(hostLines) == 2 {
				port = hostLines[1]
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
