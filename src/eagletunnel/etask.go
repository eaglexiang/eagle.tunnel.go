package eagletunnel

import "../eaglelib"

// EtASK请求的类型
const (
	EtAskLOCAL = iota
	EtAskREMOTE
	EtAskAUTH
	EtAskUNKNOWN
)

// ETAsk ET-ASK协议的实现
type ETAsk struct {
}

// Handle 处理ET-ASK请求
func (ea *ETAsk) Handle(req Request, tunnel *eaglelib.Tunnel) bool {
	return false
}

// Send 发送ET-ASK请求
func (ea *ETAsk) Send(e *NetArg) bool {
	if len(e.Args) < 1 {
		e.Reply = "is it 'ask local' or 'ask remote'?"
		return false
	}
	etAskType := parseEtAskType(e.Args[0])
	switch etAskType {
	case EtAskLOCAL:
		e.Args = e.Args[1:]
		return sendLocalReq(e)
	default:
		e.Reply = "is it 'ask local' or 'ask remote'?"
		return false
	}
}

func sendLocalReq(e *NetArg) bool {
	if len(e.Args) < 1 {
		e.Reply = "what do you want 'et ask local' to do?"
		return false
	}
	etAskType := parseEtAskType(e.Args[0])
	switch etAskType {
	case EtAskAUTH:
		e.Args = e.Args[1:]
		return sendAskAuthReq(e)
	default:
		e.Reply = "Unknown ET-ASK local type: " + e.Args[0]
		return false
	}
}

func sendAskAuthReq(e *NetArg) bool {
	if len(e.Args) < 1 {
		e.Args = append(e.Args, DefaultClientConfig())
	}
	Init(e.Args[0])
	// 当connect2Relayer成功，则说明鉴权成功
	tunnel := eaglelib.Tunnel{}
	defer tunnel.Close()
	err := connect2Relayer(&tunnel)
	if err != nil {
		e.Reply = err.Error() // 通过参数集返回具体的错误信息
		return false
	}

	if LocalUser.ID == "root" {
		e.Reply = "no local user"
	} else {
		e.Reply = "AUTH OK with local user: " + LocalUser.ID
	}
	return true
}

// parseEtAskType 将字符串转换为EtASK请求的类型
func parseEtAskType(src string) int {
	var result int
	switch src {
	case "LOCAL", "local":
		result = EtAskLOCAL
	case "REMOTE", "remote":
		result = EtAskREMOTE
	case "AUTH", "auth":
		result = EtAskAUTH
	default:
		result = EtAskUNKNOWN
	}
	return result
}

// formatEtAskType 得到EtASK请求类型对应的字符串
func formatEtAskType(src int) string {
	var result string
	switch src {
	case EtAskLOCAL:
		result = "LOCAL"
	case EtAskREMOTE:
		result = "REMOTE"
	case EtAskAUTH:
		result = "AUTH"
	default:
		result = "UNKNOWN"
	}
	return result
}
