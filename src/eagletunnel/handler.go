package eagletunnel

type Handler interface {
	handle(request Request, tunnel *Tunnel) bool
}
