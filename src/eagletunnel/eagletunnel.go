package eagletunnel

import (
	"errors"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/eaglexiang/eagle.lib.go/src"
)

// ET请求的类型
const (
	EtTCP = iota
	EtDNS
	EtLOCATION
	EtASK
	EtUNKNOWN
)

// 代理的状态
const (
	ProxyENABLE = iota
	ProxySMART
)

// protocolVersion 作为Sender使用的协议版本号
var protocolVersion, _ = eaglelib.CreateVersion("1.2")

// protocolCompatibleVersion 作为Relayer可兼容的最低版本号
var protocolCompatibleVersion, _ = eaglelib.CreateVersion("1.1")
var version, _ = eaglelib.CreateVersion("0.2")
var insideCache = sync.Map{}
var dnsRemoteCache = sync.Map{}
var hostsCache = make(map[string]string)

// EncryptKey 所有Tunnel使用的Key
var EncryptKey byte

// EagleTunnel 遵循ET协议的数据隧道
type EagleTunnel struct {
}

// Handle 处理ET请求
func (et *EagleTunnel) Handle(request Request, tunnel *eaglelib.Tunnel) (willContinue bool) {
	var result bool
	args := strings.Split(request.RequestMsgStr, " ")
	isVersionOk := et.checkVersionOfReq(args, tunnel)
	if isVersionOk {
		tunnel.EncryptLeft = true
		isUserOk := checkUserOfReq(tunnel)
		if isUserOk {
			buffer := make([]byte, 1024)
			count, _ := tunnel.ReadLeft(buffer)
			if count > 0 {
				req := string(buffer[:count])
				args := strings.Split(req, " ")
				reqType := ParseEtType(args[0])
				etReq := Request{RequestMsgStr: req}
				switch reqType {
				case EtDNS:
					ed := ETDNS{}
					ed.Handle(etReq, tunnel)
				case EtTCP:
					ett := ETTCP{}
					result = ett.Handle(etReq, tunnel)
				case EtLOCATION:
					el := ETLocation{}
					el.Handle(etReq, tunnel)
				case EtASK:
					ea := ETAsk{}
					ea.Handle(etReq, tunnel)
				default:
				}
			}
		}
	}
	return result
}

// Send 发送ET请求
func (et *EagleTunnel) Send(e *NetArg) (succeed bool) {
	var result bool
	switch e.TheType {
	case EtDNS:
		ed := ETDNS{}
		result = ed.Send(e)
	case EtTCP:
		ett := ETTCP{}
		result = ett.Send(e)
	case EtLOCATION:
		el := ETLocation{}
		result = el.Send(e)
	case EtASK:
		et := ETAsk{}
		result = et.Send(e)
	default:
	}
	return result
}

func connect2Relayer(tunnel *eaglelib.Tunnel) error {
	remoteIpe := RemoteAddr + ":" + RemotePort
	conn, err := net.DialTimeout("tcp", remoteIpe, 5*time.Second)
	if err != nil {
		return err
	}
	tunnel.Right = &conn
	tunnel.EncryptKey = EncryptKey
	err = checkVersionOfRelayer(tunnel)
	if err != nil {
		return err
	}
	tunnel.EncryptRight = true
	err = checkUserOfLocal(tunnel)
	return err
}

func checkVersionOfRelayer(tunnel *eaglelib.Tunnel) error {
	req := "eagle_tunnel " + protocolVersion.Raw + " simple"
	count, err := tunnel.WriteRight([]byte(req))
	if err != nil {
		return err
	}
	var buffer = make([]byte, 1024)
	count, err = tunnel.ReadRight(buffer)
	if err != nil {
		return err
	}
	reply := string(buffer[0:count])
	if reply != "valid valid valid" {
		replys := strings.Split(reply, " ")
		reply = ""
		for _, item := range replys {
			reply += " \"" + item + "\""
		}
		return errors.New(reply)
	}
	return err
}

func (et *EagleTunnel) checkVersionOfReq(
	headers []string,
	tunnel *eaglelib.Tunnel) (isValid bool) {
	var result bool
	if len(headers) >= 3 {
		replys := make([]string, 3)
		if headers[0] == "eagle_tunnel" {
			replys[0] = "valid"
		} else {
			replys[0] = "invalid"
		}
		versionOfReq, err := eaglelib.CreateVersion(headers[1])
		if err == nil {
			if protocolCompatibleVersion.IsSTOrE2(&versionOfReq) {
				replys[1] = "valid"
			} else {
				replys[1] = "incompatible et protocol version"
			}
		} else {
			replys[1] = err.Error()
		}
		if headers[2] == "simple" {
			replys[2] = "valid"
		} else {
			replys[2] = "invalid"
		}
		reply := replys[0] + " " + replys[1] + " " + replys[2]
		count, _ := tunnel.WriteLeft([]byte(reply))
		result = count == 17 // length of "valid valid valid"
	}
	return result
}

func checkUserOfLocal(tunnel *eaglelib.Tunnel) error {
	var err error
	if LocalUser.ID == "root" {
		return nil // no need to check
	}
	user := LocalUser.toString()
	var count int
	count, err = tunnel.WriteRight([]byte(user))
	if err != nil {
		return err
	}
	buffer := make([]byte, 1024)
	count, err = tunnel.ReadRight(buffer)
	if err != nil {
		return err
	}
	reply := string(buffer[:count])
	if reply != "valid" {
		err = errors.New(reply)
	} else {
		LocalUser.addTunnel(tunnel)
	}
	return err
}

func checkUserOfReq(tunnel *eaglelib.Tunnel) (isValid bool) {
	var result bool
	if EnableUserCheck {
		buffer := make([]byte, 1024)
		count, _ := tunnel.ReadLeft(buffer)
		if count > 0 {
			userStr := string(buffer[:count])
			addr := (*tunnel.Left).RemoteAddr()
			ip := strings.Split(addr.String(), ":")[0]
			user2Check, err := ParseEagleUser(userStr, ip)
			if err == nil {
				if user2Check.ID == "root" {
					tunnel.WriteLeft([]byte("id shouldn't be 'root'"))
				} else {
					validUser, ok := Users[user2Check.ID]
					if ok {
						err = validUser.CheckAuth(user2Check)
						if err == nil {
							reply := "valid"
							count, _ = tunnel.WriteLeft([]byte(reply))
							result = count == 5
							if result {
								validUser.addTunnel(tunnel)
							}
						} else {
							reply := err.Error()
							_, _ = tunnel.WriteLeft([]byte(reply))
						}
					} else {
						tunnel.WriteLeft([]byte("incorrent username or password"))
					}
				}
			}
		}
	} else {
		result = true
	}
	return result
}

// ParseEtType 得到字符串对应的ET请求类型
func ParseEtType(src string) int {
	var result int
	switch src {
	case "DNS":
		result = EtDNS
	case "TCP":
		result = EtTCP
	case "LOCATION":
		result = EtLOCATION
	case "ASK":
		result = EtASK
	default:
		result = EtUNKNOWN
	}
	return result
}

// FormatEtType 得到ET请求类型对应的字符串
func FormatEtType(src int) string {
	var result string
	switch src {
	case EtDNS:
		result = "DNS"
	case EtTCP:
		result = "TCP"
	case EtLOCATION:
		result = "LOCATION"
	case EtASK:
		result = "ASK"
	default:
		result = "UNKNOWN"
	}
	return result
}
