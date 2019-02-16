/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-27 08:24:42
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-02-17 00:53:56
 */

package et

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/eaglexiang/go-bytebuffer"

	mytunnel "github.com/eaglexiang/go-tunnel"
	version "github.com/eaglexiang/go-version"
)

// EtASK请求的类型
const (
	EtCheckUNKNOWN = iota
	EtCheckAuth
	EtCheckPING
	EtCheckVERSION
)

// Check ET-Check协议的实现
type Check struct {
}

// ParseEtCheckType 将字符串转换为EtCHECK请求的类型
func ParseEtCheckType(src string) int {
	switch src {
	case "AUTH", "auth":
		return EtCheckAuth
	case "PING", "ping":
		return EtCheckPING
	case "VERSION", "version":
		return EtCheckVERSION
	default:
		return EtCheckUNKNOWN
	}
}

// formatEtCheckType 得到EtCHECK请求类型对应的字符串
func formatEtCheckType(src int) string {
	switch src {
	case EtCheckAuth:
		return "AUTH"
	case EtCheckPING:
		return "PING"
	case EtCheckVERSION:
		return "VERSION"
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
	default:
		return errors.New("Check.Handle -> invalid check type")
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

// SendEtCheckAuthReq 发射 ET-CHECK-AUTH 请求
func SendEtCheckAuthReq(et *ET) string {
	// 当connect2Relayer成功，则说明鉴权成功
	tunnel := mytunnel.GetTunnel()
	defer mytunnel.PutTunnel(tunnel)
	err := et.connect2Relayer(tunnel)
	if err != nil {
		return err.Error()
	}

	if et.localUser.ID() == "null" {
		return "no local user"
	}
	return "AUTH OK with local user: " + et.localUser.ID()
}

// SendEtCheckVersionReq 发射 ET-CHECK-VERSION 请求
func SendEtCheckVersionReq(et *ET) string {
	tunnel := mytunnel.GetTunnel()
	defer mytunnel.PutTunnel(tunnel)
	err := et.connect2Relayer(tunnel)
	if err != nil {
		return err.Error()
	}

	// 告知VERSION请求
	req := FormatEtType(EtCHECK) + " " +
		formatEtCheckType(EtCheckVERSION) + " " +
		ProtocolVersion.Raw
	_, err = tunnel.WriteRight([]byte(req))
	if err != nil {
		return err.Error()
	}

	// 接受回复
	buffer := bytebuffer.GetKBBuffer()
	defer bytebuffer.PutKBBuffer(buffer)
	buffer.Length, err = tunnel.ReadRight(buffer.Buf())
	if err != nil {
		return err.Error()
	}
	reply := buffer.String()
	return reply
}

// SendEtCheckPingReq 发射ET-CHECK-PING请求
func SendEtCheckPingReq(et *ET, sig chan string) {

	start := time.Now() // 开始计时

	// 连接服务器
	tunnel := mytunnel.GetTunnel()
	defer mytunnel.PutTunnel(tunnel)
	err := et.connect2Relayer(tunnel)
	if err != nil {
		sig <- "SendEtCheckPingReq-> " + err.Error()
		return
	}

	// 告知PING请求
	req := FormatEtType(EtCHECK) + " " + formatEtCheckType(EtCheckPING)
	_, err = tunnel.WriteRight([]byte(req))
	if err != nil {
		sig <- "SendEtCheckPingReq-> " + err.Error()
		return
	}

	// 接收响应数据
	buffer := bytebuffer.GetKBBuffer()
	defer bytebuffer.PutKBBuffer(buffer)
	buffer.Length, err = tunnel.ReadRight(buffer.Buf())
	end := time.Now() // 停止计时
	if err != nil {
		sig <- err.Error()
		return
	}
	reply := buffer.String()
	if reply != "ok" {
		sig <- "SendEtCheckPingReq-> " + "invalid ping reply: " + reply
		return
	}
	duration := end.Sub(start)
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
	if versionOfReq.IsLessThan(&ProtocolCompatibleVersion) {
		reply := "the version of protocol may be incompatible"
		tunnel.WriteLeft([]byte(reply))
		return
	}
	reply := "Protocol Version OK"
	tunnel.WriteLeft([]byte(reply))
}

// Type ET子协议的类型
func (c Check) Type() int {
	return EtCHECK
}
