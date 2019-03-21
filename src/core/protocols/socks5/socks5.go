/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-04 17:56:15
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-03-17 20:09:53
 */

package socks5

import (
	"encoding/binary"
	"errors"
	"net"
	"strings"

	"github.com/eaglexiang/eagle.tunnel.go/src/logger"

	"github.com/eaglexiang/go-bytebuffer"
	mynet "github.com/eaglexiang/go-net"
	mytunnel "github.com/eaglexiang/go-tunnel"
)

// SOCKS请求的类型
const (
	SOCKSERROR = iota
	SOCKSCONNECT
	SOCKSBIND
	SOCKSUDP

	// SOCKS请求是否成功的反馈
	REPSUCCESS = iota
	REPERROR
)

var commands map[byte]command

func init() {
	commands = make(map[byte]command)
	commands[SOCKSCONNECT] = connect{}
	commands[SOCKSBIND] = bind{}
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
	if len(firstMsg) < 3 {
		return false
	}
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

func checkTunnel(tunnel *mytunnel.Tunnel) error {
	if tunnel == nil {
		return errors.New("Socks5.Handle -> tunnel is nil")
	}
	// 不接受来自公网IP的SOCKS5请求
	ipOfReq := strings.Split(tunnel.Left.RemoteAddr().String(), ":")[0]
	if !mynet.CheckPrivateIPv4(ipOfReq) {
		logger.Warning("invalid public req from ", ipOfReq)
		return errors.New("invalid public req")
	}
	return nil
}

func getCommand(tunnel *mytunnel.Tunnel,
	buffer *bytebuffer.ByteBuffer) (cmd command, err error) {
	cmd, ok := commands[buffer.Buf()[1]]
	if !ok {
		return nil, errors.New("Socks5.Handle -> invalid req")
	}
	return cmd, nil
}

func getMsgFromL(tunnel *mytunnel.Tunnel) (buffer *bytebuffer.ByteBuffer, err error) {
	buffer = bytebuffer.GetKBBuffer()
	buffer.Length, err = tunnel.ReadLeft(buffer.Buf())
	return
}

func checkVersion(e *mynet.Arg) (err error) {
	version := e.Msg[0]
	if version != '\u0005' {
		return errors.New("Socks5.Handle -> invalid socks version")
	}
	reply := "\u0005\u0000"
	count, err := e.Tunnel.WriteLeft([]byte(reply))
	if count < 2 {
		return errors.New("Scosk5.Handle -> fail to reply")
	}
	return nil
}

// Handle 处理SOCKS5请求
func (conn *Socks5) Handle(e *mynet.Arg) (err error) {
	err = checkTunnel(e.Tunnel)
	if err != nil {
		return err
	}
	err = checkVersion(e)
	if err != nil {
		return err
	}
	req, err := getMsgFromL(e.Tunnel)
	defer bytebuffer.PutKBBuffer(req)
	if err != nil {
		return err
	}
	cmd, err := getCommand(e.Tunnel, req)
	if err != nil {
		return err
	}
	return cmd.Handle(req.Cut(), e)
}

func getHost(request []byte) (host string, err error) {
	var destype = request[3]
	switch destype {
	case AddrV4:
		ip := net.IP(request[4:8])
		host = ip.String()
	case AddrDomain:
		len := request[4]
		host = string(request[5 : 5+len])
	case AddrV6:
		ip := net.IP(request[4:20])
		host = ip.String()
	default:
		logger.Warning("invalid socks req des type: ", destype)
		return "", errors.New("invalid socks req des type")
	}
	return host, nil
}

func getPort(request []byte) (port int, err error) {
	destype := request[3]
	var buffer []byte
	switch destype {
	case AddrV4:
		buffer = request[8:10]
	case AddrDomain:
		len := request[4]
		buffer = request[5+len : 7+len]
	case AddrV6:
		buffer = request[20:22]
	default:
		return 0, errors.New("getPort -> invalid destype")
	}
	return int(binary.BigEndian.Uint16(buffer)), nil
}

func getHostAndPort(request []byte) (host string, port int, err error) {
	host, err = getHost(request)
	if err == nil {
		port, err = getPort(request)
	}
	if err != nil {
		logger.Warning("fail to get host or port: ", err)
	}
	return
}
