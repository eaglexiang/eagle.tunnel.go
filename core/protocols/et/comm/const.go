package comm

// CMDType ET子协议的命令类型
type CMDType int

// ET子协议的类型
const (
	UNKNOWN CMDType = iota
	TCP
	DNS
	DNS6
	LOCATION
	CHECK
	BIND
	NEWIPE
)

// ET子协议的名字
const (
	UNKNOWNTxt  = "UNKNOWN"
	TCPTxt      = "TCP"
	DNSTxt      = "DNS"
	DNS6Txt     = "DNS6"
	LOCATIONTxt = "LOCATION"
	CHECKTxt    = "CHECK"
	BINDTxt     = "BIND"
	NEWIPETxt   = "NEWIPE"
)

// 代理的状态
const (
	ProxyENABLE = iota
	ProxySMART
	ErrorProxyStatus
)

// 代理状态对应的文本
const (
	ProxyEnableText      = "ENABLE"
	ProxySmartText       = "SMART"
	ErrorProxyStatusText = "ERROR"
)

// 域名的类型
const (
	UncertainDomain = iota
	ProxyDomain
	DirectDomain
)
