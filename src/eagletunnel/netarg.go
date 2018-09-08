package eagletunnel

type NetArg struct {
	domain  string
	ip      string
	port    int
	tunnel  *Tunnel
	user    *EagleUser
	boolObj bool
	theType int
}
