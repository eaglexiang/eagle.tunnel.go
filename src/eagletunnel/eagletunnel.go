/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-27 08:24:57
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-21 17:34:41
 */

package eagletunnel

import (
	"errors"
	"net"
	"strings"
	"time"

	mytunnel "github.com/eaglexiang/go-tunnel"
	version "github.com/eaglexiang/go-version"
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
var ProtocolVersion, _ = version.CreateVersion("1.3")

// ProtocolCompatibleVersion 作为Handler可兼容的最低协议版本号
var ProtocolCompatibleVersion, _ = version.CreateVersion("1.3")

// LocalCipherType 本地使用的加密方式
var LocalCipherType = SimpleCipherType

// EagleTunnel 遵循ET协议的数据隧道
type EagleTunnel struct {
}

// Handle 处理ET请求
func (et *EagleTunnel) Handle(request Request, tunnel *mytunnel.Tunnel) (err error) {
	args := strings.Split(request.RequestMsgStr, " ")
	err = checkHeaderOfReq(args, tunnel)
	if err != nil {
		return errors.New("EagleTunnel.Handle -> " + err.Error())
	}
	var c Cipher
	cipherType := ParseCipherType(ConfigKeyValues["cipher"])
	switch cipherType {
	case SimpleCipherType:
		c = new(SimpleCipher)
		c.SetPassword(ConfigKeyValues["data-key"])
	default:
		return errors.New("EagleTunnel.Handle -> invalid cipher type")
	}
	tunnel.EncryptLeft = c.Encrypt
	tunnel.DecryptLeft = c.Decrypt
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
func connect2Relayer(tunnel *mytunnel.Tunnel) error {
	remoteIpe := RemoteAddr + ":" + RemotePort
	conn, err := net.DialTimeout("tcp", remoteIpe, time.Second*time.Duration(Timeout))
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
	tunnel.EncryptRight = c.Encrypt
	tunnel.DecryptRight = c.Decrypt
	err = checkUserOfLocal(tunnel)
	if err != nil {
		return errors.New("connect2Relayer -> " + err.Error())
	}
	return nil
}

func checkVersionOfRelayer(tunnel *mytunnel.Tunnel) error {
	req := ConfigKeyValues["head"]
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

func checkHeaderOfReq(
	headers []string,
	tunnel *mytunnel.Tunnel) error {
	if len(headers) < 1 {
		return errors.New("checkHeaderOfReq -> nil req")
	}
	if headers[0] != ConfigKeyValues["head"] {
		return errors.New("checkHeaderOfReq -> wrong head: " + headers[0])
	}
	reply := "valid valid valid"
	count, _ := tunnel.WriteLeft([]byte(reply))
	if count != 17 {
		return errors.New("checkHeaderOfReq -> fail to reply")
	}
	return nil
}

func checkUserOfLocal(tunnel *mytunnel.Tunnel) error {
	var err error
	if LocalUser.ID == "null" {
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
	tunnel.SpeedLimiter = LocalUser.SpeedLimiter()
	return nil
}

func checkUserOfReq(tunnel *mytunnel.Tunnel) error {
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
	if user2Check.ID == "null" {
		reply := "username shouldn't be 'null'"
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
	tunnel.SpeedLimiter = validUser.SpeedLimiter()
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
