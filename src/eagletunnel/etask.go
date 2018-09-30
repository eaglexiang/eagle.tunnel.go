package eagletunnel

import (
	"strconv"
	"strings"
	"time"

	"github.com/eaglexiang/eagle.lib.go/src"
)

// EtASK请求的类型
const (
	EtAskLOCAL = iota
	EtAskREMOTE
	EtAskAUTH
	EtAskPING
	EtAskUNKNOWN
)

// ETAsk ET-ASK协议的实现
type ETAsk struct {
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
	case "PING", "ping":
		result = EtAskPING
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
	case EtAskPING:
		result = "PING"
	default:
		result = "UNKNOWN"
	}
	return result
}

// Handle 处理ET-ASK请求
func (ea *ETAsk) Handle(req Request, tunnel *eaglelib.Tunnel) bool {
	reqs := strings.Split(req.RequestMsgStr, " ")
	if len(reqs) < 2 {
		// 没有具体的ASK请求内容
		return false
	}
	etAskType := parseEtAskType(reqs[1])
	switch etAskType {
	case EtAskPING:
		handleEtAskPingReq(tunnel)
	default:
	}
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
	case EtAskPING:
		e.Args = e.Args[1:]
		return sendAskPingReq(e)
	default:
		e.Reply = "Unknown ET-ASK local type: " + e.Args[0]
		return false
	}
}

func sendAskAuthReq(e *NetArg) bool {
	if len(e.Args) < 1 {
		e.Args = append(e.Args, DefaultClientConfig())
	}
	err := Init(e.Args[0])
	if err != nil {
		e.Reply = err.Error()
		return false
	}

	// 当connect2Relayer成功，则说明鉴权成功
	tunnel := eaglelib.Tunnel{}
	defer tunnel.Close()
	err = connect2Relayer(&tunnel)
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

func sendAskPingReq(e *NetArg) bool {
	if len(e.Args) < 1 {
		e.Args = append(e.Args, DefaultClientConfig())
	}
	err := Init(e.Args[0])
	if err != nil {
		e.Reply = err.Error()
		return false
	}
	// 连接服务器
	tunnel := eaglelib.Tunnel{}
	defer tunnel.Close()
	err = connect2Relayer(&tunnel)
	if err != nil {
		e.Reply = err.Error()
		return false
	}

	addr := (*tunnel.Right).RemoteAddr()
	addrs := strings.Split(addr.String(), ":")
	e.IP = addrs[0]
	port, _ := strconv.ParseInt(addrs[1], 10, 32)
	e.Port = int(port)

	// 告知ASK请求
	req := FormatEtType(EtASK) + " " + formatEtAskType(EtAskPING)
	start := time.Now() // 开始计时
	_, err = tunnel.WriteRight([]byte(req))
	if err != nil {
		e.Reply = "when send ask: " + err.Error()
		return false
	}
	// 接收响应数据
	buffer := make([]byte, 8)
	count, err := tunnel.ReadRight(buffer)
	end := time.Now() // 停止计时
	if err != nil {
		e.Reply = "when read ask reply: " + err.Error()
		return false
	}
	reply := string(buffer[:count])
	if reply != "ok" {
		e.Reply = "invalid ping reply"
		return false
	}
	duration := end.Sub(start)
	ns := duration.Nanoseconds()
	ms := ns / 1000 / 1000
	e.Reply = strconv.FormatInt(ms, 10)
	return true
}

func handleEtAskPingReq(tunnel *eaglelib.Tunnel) {
	reply := "ok"
	tunnel.WriteLeft([]byte(reply))
}
