package eagletunnel

var RemoteAddr string
var RemotePort string
var LocalUser *EagleUser

type Sender interface {
	send(e NetArg) bool
}
