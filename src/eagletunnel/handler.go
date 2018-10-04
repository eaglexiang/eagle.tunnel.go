package eagletunnel

import (
	"../eaglelib/src"
)

// Handler 请求处理者
type Handler interface {
	Handle(request Request, tunnel *eaglelib.Tunnel) bool
}
