/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-04 17:56:15
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-21 17:33:04
 */

package eagletunnel

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"

	eagletunnel "github.com/eaglexiang/go-tunnel"
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

// Handle 处理SOCKS5请求
func (conn *Socks5) Handle(request Request, tunnel *eagletunnel.Tunnel) error {
	ipOfReq := strings.Split((*tunnel.Left).RemoteAddr().String(), ":")[0]
	if !CheckPrivateIPv4(ipOfReq) {
		// 不接受来自公网IP的SOCKS5请求
		return errors.New("Socks5.Handle -> invalid source IP type: public " + ipOfReq)
	}

	version := request.requestMsg[0]
	if version != '\u0005' {
		return errors.New("Socks5.Handle -> invalid socks version")
	}
	reply := "\u0005\u0000"
	count, err := tunnel.WriteLeft([]byte(reply))
	if err != nil {
		return errors.New("Socks5.Handle -> " + err.Error())
	}
	var buffer = make([]byte, 1024)
	count, err = tunnel.ReadLeft(buffer)
	if err != nil {
		return errors.New("Socks5.Handle -> " + err.Error())
	}
	if count < 2 {
		req := string(buffer[:count])
		return errors.New("Scosk5.Handle -> invalid socks 2nd req: " + req)
	}
	cmdType := buffer[1]
	switch cmdType {
	case SOCKSCONNECT:
		err := conn.handleTCPReq(buffer[:count], tunnel)
		if err != nil {
			return errors.New("Socks5.Handle -> " + err.Error())
		}
		return nil
	default:
		return errors.New("Socks5.Handle -> invalid socks req type")
	}
}

func (conn *Socks5) handleTCPReq(req []byte, tunnel *eagletunnel.Tunnel) error {
	ip, err := conn.getIP(req)
	if err != nil {
		return errors.New("Socks5.handleTCPReq -> " + err.Error())
	}
	port := conn.getPort(req)
	if port <= 0 {
		return errors.New("Socks5.handleTCPReq -> invalid des: " +
			ip + ":" + strconv.FormatInt(int64(port), 10))
	}
	var e = NetArg{
		IP:      ip,
		Port:    port,
		tunnel:  tunnel,
		TheType: EtTCP,
	}
	newConn := EagleTunnel{}
	err = newConn.Send(&e)
	if err != nil {
		reply := "\u0005\u0001\u0000\u0001\u0000\u0000\u0000\u0000\u0000\u0000"
		tunnel.WriteLeft([]byte(reply))
		return errors.New("Socks5.handleTCPReq -> " + err.Error())
	}
	reply := "\u0005\u0000\u0000\u0001\u0000\u0000\u0000\u0000\u0000\u0000"
	_, err = tunnel.WriteLeft([]byte(reply))
	if err != nil {
		return errors.New("Socks5.handleTCPReq -> " + err.Error())
	}
	return nil
}

func (conn *Socks5) getIP(request []byte) (string, error) {
	var destype = request[3]
	switch destype {
	case 1:
		ip := fmt.Sprintf("%d.%d.%d.%d", request[4], request[5], request[6], request[7])
		return ip, nil
	case 3:
		len := request[4]
		domain := string(request[5 : 5+len])
		newConn := EagleTunnel{}
		e := NetArg{
			domain:  domain,
			TheType: EtDNS,
		}
		err := newConn.Send(&e)
		if err != nil {
			return "", errors.New("Socks5.getIP -> " + err.Error())
		}
		return e.IP, nil
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
