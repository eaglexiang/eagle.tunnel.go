/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-04 14:30:39
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-02-21 16:18:54
 */

package httpproxy

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	mynet "github.com/eaglexiang/go-net"
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

// Match 判断是否匹配
func (conn *HTTPProxy) Match(firstMsg []byte) bool {
	args := strings.Split(string(firstMsg), " ")
	switch args[0] {
	case "OPTIONS", "HEAD", "GET", "POST", "PUT", "DELETE", "TRACE", "CONNECT":
		return true
	default:
		return false
	}
}

// Handle 处理HTTPProxy请求
func (conn *HTTPProxy) Handle(e *mynet.Arg) error {
	if e.Tunnel == nil {
		return errors.New("HTTPProxy.Handle -> tunnel is nil")
	}

	// 不接受来自公网IP的HTTP代理请求
	ipOfReq := strings.Split((*e.Tunnel.Left).RemoteAddr().String(), ":")[0]
	if !mynet.CheckPrivateIPv4(ipOfReq) {
		return errors.New("HTTPProxy.Handle -> invalid source ip type: public " +
			ipOfReq)
	}

	reqStr := string(e.Msg)
	reqType, host, _port := dismantle(reqStr)
	port, err := strconv.Atoi(_port)
	if err != nil {
		return err
	}
	if host == "" || port <= 0 {
		return errors.New("invalid des: " + host + ":" + _port)
	}
	e.Host = host + ":" + _port
	// reply http proxy req
	switch reqType {
	case HTTPCONNECT:
		re443 := "HTTP/1.1 200 Connection Established\r\n\r\n"
		_, err = e.Tunnel.WriteLeft([]byte(re443))
		if err != nil {
			return errors.New("HTTPProxy.Handle -> " +
				err.Error())
		}
		return nil
	case HTTPOTHERS:
		e.Delegates = append(e.Delegates, func() {
			newReq := createNewRequest(reqStr)
			_, err = e.Tunnel.WriteRight([]byte(newReq))
		})
		return nil
	default:
		return errors.New("invalid HTTP type: " + reqStr)
	}
}

// Name 名字
func (conn *HTTPProxy) Name() string {
	return "HTTP"
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
