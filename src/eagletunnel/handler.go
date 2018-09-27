package eagletunnel

import "../eaglelib"

// Handler 请求处理者
type Handler interface {
	handle(request Request, tunnel *eaglelib.Tunnel) bool
}
