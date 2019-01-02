/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-27 08:24:57
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-02 19:33:30
 */

package eagletunnel

import (
	"errors"
	"net"
	"strings"

	"../eaglelib/src"
)

// ET请求的类型
const (
	EtTCP = iota
	EtDNS
	EtLOCATION
	EtCHECK
	EtUNKNOWN
)

// 代理的状态
const (
	ProxyENABLE = iota
	ProxySMART
)

// ProtocolVersion 作为Sender使用的协议版本号
var ProtocolVersion, _ = eaglelib.CreateVersion("1.3")

// ProtocolCompatibleVersion 作为Handler可兼容的最低协议版本号
var ProtocolCompatibleVersion, _ = eaglelib.CreateVersion("1.3")

// LocalCipherType 本地使用的加密方式
var LocalCipherType = SimpleCipherType

// EagleTunnel 遵循ET协议的数据隧道
type EagleTunnel struct {
}

// Handle 处理ET请求
func (et *EagleTunnel) Handle(request Request, tunnel *eaglelib.Tunnel) (keepAlive bool) {
	args := strings.Split(request.RequestMsgStr, " ")
	_, cipherType := et.checkHeaderOfReq(args, tunnel)
	var c Cipher
	switch cipherType {
	case SimpleCipherType:
		c = new(SimpleCipher)
		c.SetPassword(ConfigKeyValues["data-key"])
	default:
		return false
	}
	tunnel.Encrypt = c.Encrypt
	tunnel.Decrypt = c.Decrypt
	tunnel.EncryptLeft = true
	isUserOk := checkUserOfReq(tunnel)
	if !isUserOk {
		return false
	}
	buffer := make([]byte, 1024)
	count, err := tunnel.ReadLeft(buffer)
	if err != nil {
		return false
	}
	req := string(buffer[:count])
	args = strings.Split(req, " ")
	reqType := ParseEtType(args[0])
	etReq := Request{RequestMsgStr: req}
	switch reqType {
	case EtDNS:
		ed := ETDNS{}
		ed.Handle(etReq, tunnel)
	case EtTCP:
		ett := ETTCP{}
		return ett.Handle(etReq, tunnel) // 只有TCP请求有可能需要保持连接
	case EtLOCATION:
		el := ETLocation{}
		el.Handle(etReq, tunnel)
	case EtCHECK:
		ec := ETCheck{}
		ec.Handle(etReq, tunnel)
	default:
	}
	return false
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
		el.Send(e)
	default:
	}
	return result
}

// connect2Relayer 连接到下一个Relayer，完成版本校验和用户校验两个步骤
func connect2Relayer(tunnel *eaglelib.Tunnel) error {
	remoteIpe := RemoteAddr + ":" + RemotePort
	// conn, err := net.DialTimeout("tcp", remoteIpe, 5*time.Second)
	conn, err := net.Dial("tcp", remoteIpe)
	if err != nil {
		return err
	}
	tunnel.Right = &conn
	err = checkVersionOfRelayer(tunnel)
	if err != nil {
		return err
	}
	var c Cipher
	switch LocalCipherType {
	case SimpleCipherType:
		c = new(SimpleCipher)
		key, found := ConfigKeyValues["data-key"]
		if !found {
			panic("data-key not found")
		}
		err = c.SetPassword(key)
		if err != nil {
			return err
		}
	default:
		return errors.New("unknown cipher type")
	}
	tunnel.Encrypt = c.Encrypt
	tunnel.Decrypt = c.Decrypt
	tunnel.EncryptRight = true
	err = checkUserOfLocal(tunnel)
	return err
}

func checkVersionOfRelayer(tunnel *eaglelib.Tunnel) error {
	req := ConfigKeyValues["head"] + " " + ProtocolVersion.Raw + " simple"
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

func (et *EagleTunnel) checkHeaderOfReq(
	headers []string,
	tunnel *eaglelib.Tunnel) (isValid bool, cipherType int) {
	if len(headers) < 3 {
		return false, UnknownCipherType
	}
	if headers[0] != ConfigKeyValues["head"] {
		return false, UnknownCipherType
	}
	versionOfReq, err := eaglelib.CreateVersion(headers[1])
	if err != nil {
		return false, UnknownCipherType
	}
	if !ProtocolCompatibleVersion.IsLTOrE2(&versionOfReq) {
		return false, UnknownCipherType
	}
	theType := ParseCipherType(headers[2])
	if theType == UnknownCipherType {
		return false, UnknownCipherType
	}
	reply := "valid valid valid"
	count, _ := tunnel.WriteLeft([]byte(reply))
	return count == 17, theType // length of "valid valid valid"
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
	if !EnableUserCheck {
		return true
	}
	// 接收用户信息
	buffer := make([]byte, 1024)
	count, err := tunnel.ReadLeft(buffer)
	if err != nil {
		return false
	}
	userStr := string(buffer[:count])
	addr := (*tunnel.Left).RemoteAddr()
	ip := strings.Split(addr.String(), ":")[0]
	user2Check, err := ParseReqUser(userStr, ip)
	if err != nil {
		tunnel.WriteLeft([]byte(err.Error()))
		return false
	}
	if user2Check.ID == "root" {
		tunnel.WriteLeft([]byte("username shouldn't be 'root'"))
		return false
	}
	validUser, ok := Users[user2Check.ID]
	if !ok {
		// 找不到该用户
		tunnel.WriteLeft([]byte("incorrent username or password"))
		return false
	}
	err = validUser.CheckAuth(user2Check)
	if err != nil {
		reply := err.Error()
		tunnel.WriteLeft([]byte(reply))
		return false
	}
	reply := "valid"
	count, _ = tunnel.WriteLeft([]byte(reply))
	ok = count == 5
	if !ok {
		// 发送响应失败
		return false
	}
	validUser.addTunnel(tunnel)
	return true
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
	case "CHECK":
		result = EtCHECK
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
	case EtCHECK:
		result = "CHECK"
	default:
		result = "UNKNOWN"
	}
	return result
}
