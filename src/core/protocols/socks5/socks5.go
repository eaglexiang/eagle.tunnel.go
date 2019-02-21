/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-04 17:56:15
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-02-21 16:19:01
 */

package socks5

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
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
func (conn *Socks5) Handle(e *mynet.Arg) error {
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
	if err != nil {
		return errors.New("Socks5.Handle -> " + err.Error())
	}
	buffer := bytebuffer.GetKBBuffer()
	defer bytebuffer.PutKBBuffer(buffer)
	buffer.Length, err = e.Tunnel.ReadLeft(buffer.Buf())
	if err != nil {
		return errors.New("Socks5.Handle -> " + err.Error())
	}
	if count < 2 {
		req := buffer.String()
		return errors.New("Scosk5.Handle -> invalid socks 2nd req: " + req)
	}
	cmdType := buffer.Buf()[1]
	switch cmdType {
	case SOCKSCONNECT:
		err := conn.handleTCPReq(buffer.Cut(), e)
		if err != nil {
			return errors.New("Socks5.Handle -> " + err.Error())
		}
		return nil
	default:
		return errors.New("Socks5.Handle -> invalid socks req type")
	}
}

func (conn *Socks5) handleTCPReq(req []byte, e *mynet.Arg) error {
	ip, err := conn.getHost(req)
	if err != nil {
		return errors.New("Socks5.handleTCPReq -> " + err.Error())
	}
	_port := conn.getPort(req)
	port := strconv.FormatInt(int64(_port), 10)
	if _port <= 0 {
		return errors.New("Socks5.handleTCPReq -> invalid des: " +
			ip + ":" + port)
	}
	e.Host = ip + ":" + port
	reply := "\u0005\u0000\u0000\u0001\u0000\u0000\u0000\u0000\u0000\u0000"
	_, err = e.Tunnel.WriteLeft([]byte(reply))
	if err != nil {
		return errors.New("Socks5.handleTCPReq -> " + err.Error())
	}
	return nil
}

func (conn *Socks5) getHost(request []byte) (string, error) {
	var destype = request[3]
	switch destype {
	case 1:
		ip := fmt.Sprintf("%d.%d.%d.%d", request[4], request[5], request[6], request[7])
		return ip, nil
	case 3:
		len := request[4]
		domain := string(request[5 : 5+len])
		return domain, nil
	default:
		return "", errors.New("Socks5.getIP -> invalid socks req des type: " +
			strconv.FormatInt(int64(destype), 10))
	}
}

func (conn *Socks5) getPort(request []byte) int {
	destype := request[3]
	var port int16
	var buffer []byte
	var err error
	switch destype {
	case 1:
		buffer = request[8:10]
	case 3:
		len := request[4]
		buffer = request[5+len : 7+len]
	default:
		buffer = make([]byte, 0)
		err = errors.New("invalid destype")
	}
	if err == nil {
		port = int16(binary.BigEndian.Uint16(buffer))
	}
	return int(port)
}
