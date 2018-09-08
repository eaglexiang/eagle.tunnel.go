package eagletunnel

import (
	"strings"
)

const (
	Unknown = iota
	SOCKS
	HTTP_PROXY
	EAGLE_TUNNEL
)

type Request struct {
	requestMsg    []byte
	RequestMsgStr string
}

func (request *Request) getType() int {
	var result int
	if request.requestMsg[0] == 5 {
		result = SOCKS
	} else {
		request.RequestMsgStr = string(request.requestMsg[:])
		args := strings.Split(request.RequestMsgStr, " ")
		if len(args) >= 2 {
			switch args[0] {
			case "OPTIONS", "HEAD", "GET", "POST", "PUT", "DELETE", "TRACE", "CONNECT":
				result = HTTP_PROXY
			case "eagle_tunnel":
				result = EAGLE_TUNNEL
			default:
				result = Unknown
			}
		}
	}
	return result
}
