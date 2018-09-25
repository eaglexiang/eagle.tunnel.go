package eagletunnel

// Handler 请求处理者
type Handler interface {
	handle(request Request, tunnel *Tunnel) bool
}
