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
	if request.requestMsg[0] == 5 {
		return SOCKSReq
	}
	request.RequestMsgStr = string(request.requestMsg[:])
	args := strings.Split(request.RequestMsgStr, " ")
	switch args[0] {
	case "OPTIONS", "HEAD", "GET", "POST", "PUT", "DELETE", "TRACE", "CONNECT":
		return HTTPProxyReq
	case ConfigKeyValues["head"]:
		return EagleTunnelReq
	default:
		return UnknownReq
	}
}
