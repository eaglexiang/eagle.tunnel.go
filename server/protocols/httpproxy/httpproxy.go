/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-04 14:30:39
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-08-25 20:21:11
 */

package httpproxy

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/eaglexiang/go-logger"
	mynet "github.com/eaglexiang/go-net"
	"github.com/eaglexiang/go-tunnel"
)

// HTTP请求的类型
const (
	HTTPCONNECT = iota
	HTTPOTHERS
	HTTPERROR
)

const re443 = "HTTP/1.1 200 Connection Established\r\n\r\n"

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

func checkTunnel(t *tunnel.Tunnel) error {
	if t == nil {
		return errors.New("tunnel is nil")
	}
	// 不接受来自公网IP的HTTP代理请求
	ipOfReq := mynet.GetIPOfConnRemote(t.Left())
	if !mynet.IsPrivateIPv4(ipOfReq) {
		logger.Warning("invalid public req from ", ipOfReq)
		return errors.New("invlaid public req")
	}
	return nil
}

// Handle 处理HTTPProxy请求
func (conn *HTTPProxy) Handle(e *mynet.Arg) error {
	err := checkTunnel(e.Tunnel)
	if err != nil {
		return err
	}
	e.TheType = int(mynet.CONNECT)
	reqStr := string(e.Msg)
	reqType, host, port := dismantle(reqStr)
	e.Host = host + ":" + port
	switch reqType {
	case HTTPCONNECT:
		conn.handleConnect(e)
	case HTTPOTHERS:
		conn.handleOthers(e, reqStr)
	default:
		logger.Warning("invalid http type: ", reqStr)
		err = errors.New("invalid HTTP type")
	}
	return err
}

func (conn *HTTPProxy) handleConnect(e *mynet.Arg) {
	e.Delegates = append(e.Delegates, func() bool {
		err := e.Tunnel.WriteLeftStr(re443)
		return err == nil
	})
}

func (conn *HTTPProxy) handleOthers(e *mynet.Arg, reqStr string) {
	e.Delegates = append(e.Delegates, func() bool {
		newReq := createNewRequest(reqStr)
		err := e.Tunnel.WriteRightStr(newReq)
		return err == nil
	})
}

// Name 名字
func (conn *HTTPProxy) Name() string {
	return "HTTP"
}

func dismantle(request string) (reqType int, host string, port string) {
	lines := strings.Split(request, "\r\n")
	args := strings.Split(lines[0], " ")
	if len(args) < 3 {
		return HTTPERROR, "", ""
	}

	// 获取HTTP请求类型
	reqType = getReqType(args[0])
	if reqType == HTTPERROR {
		return HTTPERROR, "", ""
	}

	// 补全协议头
	args[1] = completeProtocolHeader(args[1], reqType)
	url, err := url.Parse(args[1])
	if err != nil {
		return reqType, "", ""
	}
	// 获取host与port
	host = url.Host
	// url.Host有可能包含了端口号
	hostElements := strings.Split(host, ":")
	host = hostElements[0]
	if len(hostElements) == 2 {
		port = hostElements[1]
	} else if len(hostElements) > 2 {
		return HTTPERROR, "", ""
	}
	if port != "" {
		return reqType, host, port
	}
	// url.Host未包含端口号
	port = url.Port()
	if port != "" {
		return reqType, host, port
	}
	// 无端口号，则使用协议默认端口
	switch url.Scheme {
	case "https":
		return reqType, host, "443"
	case "http":
		return reqType, host, "80"
	default:
		return reqType, host, "80"
	}
}

// completeProtocolHeader 补全协议头
func completeProtocolHeader(path string, reqType int) string {
	if !strings.HasPrefix(path, "http://") &&
		!strings.HasPrefix(path, "https://") {
		switch reqType {
		case HTTPCONNECT:
			path = "https://" + path
		case HTTPOTHERS:
			path = "http://" + path
		}
	}
	return path
}

func getReqType(reqType string) int {
	switch reqType {
	case "CONNECT":
		return HTTPCONNECT
	case "OPTIONS", "HEAD", "GET", "POST", "PUT", "DELETE", "TRACE":
		return HTTPOTHERS
	default:
		return HTTPERROR
	}
}

func createNewRequest(oldRequest string) string {
	lines := strings.Split(oldRequest, "\r\n")
	firstLine := lines[0]
	argsOfFirstLine := strings.Split(firstLine, " ")
	if len(argsOfFirstLine) != 3 {
		return ""
	}
	u, err := url.Parse(argsOfFirstLine[1])
	if err != nil {
		return ""
	}
	path := u.Path
	if u.RawQuery != "" {
		path += "?" + u.RawQuery
	}
	argsOfFirstLine[1] = path
	newFirstLine := fmt.Sprintf("%s %s %s",
		argsOfFirstLine[0],
		argsOfFirstLine[1],
		argsOfFirstLine[2])
	newReq := newFirstLine
	for _, line := range lines[1:] {
		if strings.HasPrefix(line, "Proxy-Connection:") {
			connection := strings.TrimPrefix(line, "Proxy-Connection:")
			newReq += "\r\nConnection:" + connection
		} else {
			newReq += "\r\n" + line
		}
	}
	return newReq
}
