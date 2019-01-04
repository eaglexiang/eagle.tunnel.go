/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-27 08:24:57
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-04 18:07:56
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
	EtUNKNOWN = iota
	EtTCP
	EtDNS
	EtLOCATION
	EtCHECK
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
func (et *EagleTunnel) Handle(request Request, tunnel *eaglelib.Tunnel) (err error) {
	args := strings.Split(request.RequestMsgStr, " ")
	_, cipherType := et.checkHeaderOfReq(args, tunnel)
	var c Cipher
	switch cipherType {
	case SimpleCipherType:
		c = new(SimpleCipher)
		c.SetPassword(ConfigKeyValues["data-key"])
	default:
		return errors.New("EagleTunnel.Handle -> invalid cipher type")
	}
	tunnel.Encrypt = c.Encrypt
	tunnel.Decrypt = c.Decrypt
	tunnel.EncryptLeft = true
	err = checkUserOfReq(tunnel)
	if err != nil {
		return errors.New("EagleTunnel.Handle -> " + err.Error())
	}
	buffer := make([]byte, 1024)
	count, err := tunnel.ReadLeft(buffer)
	if err != nil {
		return errors.New("EagleTunnel.Handle -> " + err.Error())
	}
	req := string(buffer[:count])
	args = strings.Split(req, " ")
	reqType := ParseEtType(args[0])
	etReq := Request{RequestMsgStr: req}
	switch reqType {
	case EtDNS:
		ed := ETDNS{}
		ed.Handle(etReq, tunnel)
		return errors.New("no need to continue")
	case EtTCP:
		// 只有TCP请求有可能需要保持连接
		ett := ETTCP{}
		err = ett.Handle(etReq, tunnel)
		return err
	case EtLOCATION:
		el := ETLocation{}
		err = el.Handle(etReq, tunnel)
		if err != nil {
			return errors.New("EagleTunnel.Handle -> " + err.Error())
		}
		return errors.New("no need to continue")
	case EtCHECK:
		ec := ETCheck{}
		ec.Handle(etReq, tunnel)
		return errors.New("no need to continue")
	default:
		return errors.New("EagleTunnel.Handle -> invalid et req type, user-check may be wrong")
	}
}

// Send 发送ET请求
func (et *EagleTunnel) Send(e *NetArg) error {
	switch e.TheType {
	case EtDNS:
		ed := ETDNS{}
		err := ed.Send(e)
		if err != nil {
			return errors.New("EagleTunnel.Send -> " + err.Error())
		}
		return nil
	case EtTCP:
		ett := ETTCP{}
		err := ett.Send(e)
		if err != nil {
			return errors.New("EagleTunnel.Send -> " + err.Error())
		}
		return nil
	case EtLOCATION:
		el := ETLocation{}
		err := el.Send(e)
		if err != nil {
			return errors.New("EagleTunnel.Send -> " + err.Error())
		}
		return nil
	default:
		return errors.New("EagleTunnel.Send -> invalid et req type")
	}
}

// connect2Relayer 连接到下一个Relayer，完成版本校验和用户校验两个步骤
func connect2Relayer(tunnel *eaglelib.Tunnel) error {
	remoteIpe := RemoteAddr + ":" + RemotePort
	// conn, err := net.DialTimeout("tcp", remoteIpe, 5*time.Second)
	conn, err := net.Dial("tcp", remoteIpe)
	if err != nil {
		return errors.New("connect2Relayer -> " + err.Error())
	}
	tunnel.Right = &conn
	err = checkVersionOfRelayer(tunnel)
	if err != nil {
		return errors.New("connect2Relayer -> " + err.Error())
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
			return errors.New("connect2Relayer -> " + err.Error())
		}
	default:
		return errors.New("connect2Relayer -> unknown cipher type")
	}
	tunnel.Encrypt = c.Encrypt
	tunnel.Decrypt = c.Decrypt
	tunnel.EncryptRight = true
	err = checkUserOfLocal(tunnel)
	if err != nil {
		return errors.New("connect2Relayer -> " + err.Error())
	}
	LocalUser.addTunnel(tunnel)
	return nil
}

func checkVersionOfRelayer(tunnel *eaglelib.Tunnel) error {
	req := ConfigKeyValues["head"] + " " + ProtocolVersion.Raw + " simple"
	count, err := tunnel.WriteRight([]byte(req))
	if err != nil {
		return errors.New("checkVersionOfRelayer -> " + err.Error())
	}
	var buffer = make([]byte, 1024)
	count, err = tunnel.ReadRight(buffer)
	if err != nil {
		return errors.New("checkVersionOfRelayer -> " + err.Error())
	}
	reply := string(buffer[0:count])
	if reply != "valid valid valid" {
		return errors.New("checkVersionOfRelayer -> " + reply)
	}
	return nil
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
		return errors.New("checkUserOfLocal -> " + err.Error())
	}
	buffer := make([]byte, 1024)
	count, err = tunnel.ReadRight(buffer)
	if err != nil {
		return errors.New("checkUserOfLocal -> " + err.Error())
	}
	reply := string(buffer[:count])
	if reply != "valid" {
		return errors.New("checkUserOfLocal -> " + reply)
	}
	return nil
}

func checkUserOfReq(tunnel *eaglelib.Tunnel) error {
	if !EnableUserCheck {
		return nil
	}
	// 接收用户信息
	buffer := make([]byte, 1024)
	count, err := tunnel.ReadLeft(buffer)
	if err != nil {
		return errors.New("checkUserOfReq -> " + err.Error())
	}
	userStr := string(buffer[:count])
	addr := (*tunnel.Left).RemoteAddr()
	ip := strings.Split(addr.String(), ":")[0]
	user2Check, err := ParseReqUser(userStr, ip)
	if err != nil {
		tunnel.WriteLeft([]byte(err.Error()))
		return errors.New("checkUserOfReq -> " + err.Error())
	}
	if user2Check.ID == "root" {
		reply := "username shouldn't be 'root'"
		tunnel.WriteLeft([]byte(reply))
		return errors.New("checkUserOfReq -> " + reply)
	}
	validUser, ok := Users[user2Check.ID]
	if !ok {
		// 找不到该用户
		reply := "incorrent username or password"
		tunnel.WriteLeft([]byte(reply))
		return errors.New("checkUserOfReq -> username not found: " + user2Check.ID)
	}
	err = validUser.CheckAuth(user2Check)
	if err != nil {
		reply := err.Error()
		tunnel.WriteLeft([]byte(reply))
		return errors.New("checkUserOfReq -> " + err.Error())
	}
	reply := "valid"
	count, err = tunnel.WriteLeft([]byte(reply))
	if err != nil {
		return errors.New("checkUserOfReq -> " + err.Error())
	}
	if count != 5 {
		// 发送响应失败
		return errors.New("checkUserOfReq -> wrong reply")
	}
	validUser.addTunnel(tunnel)
	return nil
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
