package eagletunnel

import (
	"strings"
)

// 请求类型
const (
	UnknownReq = iota
	SOCKSReq
	HTTPProxyReq
	EagleTunnelReq
)

// Request 请求，包含字节数据和对应的字符串
type Request struct {
	requestMsg    []byte
	RequestMsgStr string
}

func (request *Request) getType() int {
	var result int
	if request.requestMsg[0] == 5 {
		result = SOCKSReq
	} else {
		request.RequestMsgStr = string(request.requestMsg[:])
		args := strings.Split(request.RequestMsgStr, " ")
		if len(args) >= 2 {
			switch args[0] {
			case "OPTIONS", "HEAD", "GET", "POST", "PUT", "DELETE", "TRACE", "CONNECT":
				result = HTTPProxyReq
			case "eagle_tunnel":
				result = EagleTunnelReq
			default:
				result = UnknownReq
			}
		}
	}
	return result
}
