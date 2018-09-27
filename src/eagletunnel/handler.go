package eagletunnel

import (
	"github.com/eaglexiang/eagle.lib.go/src"
)

// Handler 请求处理者
type Handler interface {
	Handle(request Request, tunnel *eaglelib.Tunnel) bool
}
