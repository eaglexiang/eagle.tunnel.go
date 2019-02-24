/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-04 17:56:15
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-02-24 18:31:18
 */

package socks5

import (
	"errors"
	"strings"

	"github.com/eaglexiang/go-bytebuffer"
	mynet "github.com/eaglexiang/go-net"
)

// SOCKS请求的类型
const (
	SOCKSERROR = iota
	SOCKSCONNECT
	SOCKSBIND
	SOCKSUDP
)

var commands map[byte]command

func init() {
	commands = make(map[byte]command)
	commands[SOCKSCONNECT] = connect{}
}

// command SOCKS5的子命令
type command interface {
	Handle([]byte, *mynet.Arg) error
}

// Socks5 Socks5协议的实现
type Socks5 struct {
}

// Match 匹配业务
func (conn *Socks5) Match(firstMsg []byte) bool {
	version := firstMsg[0]
	if version == '\u0005' {
		return true
	}
	return false
}

// Name 名字
func (conn *Socks5) Name() string {
	return "SOCKS"
}

// Handle 处理SOCKS5请求
func (conn *Socks5) Handle(e *mynet.Arg) (err error) {
	if e.Tunnel == nil {
		return errors.New("Socks5.Handle -> tunnel is nil")
	}

	// 不接受来自公网IP的SOCKS5请求
	ipOfReq := strings.Split((*e.Tunnel.Left).RemoteAddr().String(), ":")[0]
	if !mynet.CheckPrivateIPv4(ipOfReq) {
		return errors.New("Socks5.Handle -> invalid source IP type: public " + ipOfReq)
	}

	version := e.Msg[0]
	if version != '\u0005' {
		return errors.New("Socks5.Handle -> invalid socks version")
	}
	reply := "\u0005\u0000"
	count, err := e.Tunnel.WriteLeft([]byte(reply))
	if count < 2 {
		return errors.New("Scosk5.Handle -> fail to reply")
	}
	buffer := bytebuffer.GetKBBuffer()
	defer bytebuffer.PutKBBuffer(buffer)
	buffer.Length, err = e.Tunnel.ReadLeft(buffer.Buf())
	if err != nil {
		return errors.New("Socks5.Handle -> " + err.Error())
	}
	cmdType := buffer.Buf()[1]
	command, ok := commands[cmdType]
	if ok {
		err = command.Handle(buffer.Cut(), e)
	} else {
		err = errors.New("invalid socks req type")
	}
	if err != nil {
		return errors.New("Socks5.Handle -> " + err.Error())
	}
	return nil
}
