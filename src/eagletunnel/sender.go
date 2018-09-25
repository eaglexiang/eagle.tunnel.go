package eagletunnel

// RemoteAddr 上级Relayer地址
var RemoteAddr string

// RemotePort 上级Relayer端口
var RemotePort string

// Sender 请求发送者
type Sender interface {
	send(e NetArg) bool
}
