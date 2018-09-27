package eagletunnel

import "../eaglelib"

// NetArg 服务内部需要的参数集合
type NetArg struct {
	domain  string
	ip      string
	port    int
	tunnel  *eaglelib.Tunnel
	user    *EagleUser
	boolObj bool
	TheType int
	Reply   string
	Args    []string
}
