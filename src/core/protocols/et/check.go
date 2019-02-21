/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-27 08:24:42
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-02-21 18:19:29
 */

package et

import (
	"errors"
	"strconv"
	"strings"
	"time"

	myuser "github.com/eaglexiang/go-user"

	mytunnel "github.com/eaglexiang/go-tunnel"
	version "github.com/eaglexiang/go-version"
)

// EtASK请求的类型
const (
	EtCheckUNKNOWN = iota
	EtCheckAUTH
	EtCheckPING
	EtCheckVERSION
	EtCheckUSERS
)

// Check ET-Check协议的实现
type Check struct {
	validUsers map[string]*myuser.User
}

// ParseEtCheckType 将字符串转换为EtCHECK请求的类型
func ParseEtCheckType(src string) int {
	switch src {
	case "AUTH", "auth":
		return EtCheckAUTH
	case "PING", "ping":
		return EtCheckPING
	case "VERSION", "version":
		return EtCheckVERSION
	case "USERS", "users":
		return EtCheckUSERS
	default:
		return EtCheckUNKNOWN
	}
}

// formatEtCheckType 得到EtCHECK请求类型对应的字符串
func formatEtCheckType(src int) string {
	switch src {
	case EtCheckAUTH:
		return "AUTH"
	case EtCheckPING:
		return "PING"
	case EtCheckVERSION:
		return "VERSION"
	case EtCheckUSERS:
		return "USERS"
	default:
		return "UNKNOWN"
	}
}

// Handle 处理ET-Check请求
func (c Check) Handle(req string, tunnel *mytunnel.Tunnel) error {
	reqs := strings.Split(req, " ")
	if len(reqs) < 2 {
		return errors.New("Check.Handle -> no value for req")
	}
	theType := ParseEtCheckType(reqs[1])
	switch theType {
	case EtCheckPING:
		handleEtCheckPingReq(tunnel)
	case EtCheckVERSION:
		handleEtCheckVersionReq(tunnel, reqs)
	case EtCheckUSERS:
		c.handleEtCheckUsersReq(tunnel)
	default:
		return errors.New("Check.Handle -> invalid check type: " +
			reqs[1])
	}
	return nil
}

// Match 判断是否匹配
func (c Check) Match(req string) bool {
	args := strings.Split(req, " ")
	if args[0] == "CHECK" {
		return true
	}
	return false
}

// Type ET子协议的类型
func (c Check) Type() int {
	return EtCHECK
}

// SendEtCheckAuthReq 发射 ET-CHECK-AUTH 请求
func SendEtCheckAuthReq(et *ET) string {
	// null代表未启用本地用户
	if et.localUser.ID() == "null" {
		return "no local user"
	}

	// 当connect2Relayer成功，则说明鉴权成功
	tunnel := mytunnel.GetTunnel()
	defer mytunnel.PutTunnel(tunnel)
	err := et.connect2Relayer(tunnel)
	if err != nil {
		return err.Error()
	}

	return "AUTH OK with local user: " + et.localUser.ID()
}

// SendEtCheckVersionReq 发射 ET-CHECK-VERSION 请求
func SendEtCheckVersionReq(et *ET) string {
	req := FormatEtType(EtCHECK) + " " +
		formatEtCheckType(EtCheckVERSION) + " " +
		ProtocolVersion.Raw
	reply := sendQueryReq(et, req)
	return reply
}

// SendEtCheckPingReq 发射ET-CHECK-PING请求
func SendEtCheckPingReq(et *ET, sig chan string) {

	start := time.Now() // 开始计时

	req := FormatEtType(EtCHECK) + " " + formatEtCheckType(EtCheckPING)
	reply := sendQueryReq(et, req)
	if reply != "ok" {
		sig <- "SendEtCheckPingReq-> invalid reply: " + reply
		return
	}

	duration := time.Since(start)
	ns := duration.Nanoseconds()
	ms := ns / 1000 / 1000
	sig <- strconv.FormatInt(ms, 10)
	return
}

func handleEtCheckPingReq(tunnel *mytunnel.Tunnel) {
	reply := "ok"
	tunnel.WriteLeft([]byte(reply))
}

func handleEtCheckVersionReq(tunnel *mytunnel.Tunnel, reqs []string) {
	if len(reqs) < 3 {
		reply := "no protocol version value"
		tunnel.WriteLeft([]byte(reply))
		return
	}
	versionOfReq, err := version.CreateVersion(reqs[2])
	if err != nil {
		reply := err.Error()
		tunnel.WriteLeft([]byte(reply))
		return
	}
	if versionOfReq.IsLessThan(ProtocolCompatibleVersion) {
		reply := "the version of protocol may be incompatible"
		tunnel.WriteLeft([]byte(reply))
		return
	}
	reply := "Protocol Version OK"
	tunnel.WriteLeft([]byte(reply))
}

// SendEtCheckUsersReq 发射 ET-CHECK-USERS 请求
func SendEtCheckUsersReq(et *ET) string {
	req := FormatEtType(EtCHECK) + " " +
		formatEtCheckType(EtCheckUSERS)
	reply := sendQueryReq(et, req)
	return reply
}

func (c Check) handleEtCheckUsersReq(tunnel *mytunnel.Tunnel) {
	var reply string
	for _, user := range c.validUsers {
		line := user.ID() + ": " + user.Count()
		reply += line + "\n"
	}
	tunnel.WriteLeft([]byte(reply))
}
