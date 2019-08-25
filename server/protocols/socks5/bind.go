/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-02-24 18:40:56
 * @LastEditTime: 2019-08-25 20:21:53
 */

package socks5

import (
	"encoding/binary"
	"errors"
	"net"
	"strconv"
	"strings"

	"github.com/eaglexiang/go-bytebuffer"
	mynet "github.com/eaglexiang/go-net"
)

type bind struct {
}

func (b bind) Handle(req []byte, e *mynet.Arg) error {
	host, _port, err := getHostAndPort(req)
	if err != nil {
		return errors.New("connect.Handle -> " +
			err.Error())
	}
	port := strconv.FormatInt(int64(_port), 10)
	e.Host = host + ":" + port
	e.TheType = int(mynet.BIND)
	// 根据BIND的结果对客户端进行反馈
	e.Delegates = append(e.Delegates, func() bool {
		var reply []byte
		defer func() {
			b := bytebuffer.GetBytesBuffer(reply)
			defer bytebuffer.PutBuffer(b)

			e.Tunnel.WriteLeft(b)
		}()

		if e.TheType != 0 {
			reply = []byte{5, byte(REPERROR), 0, 1, 0, 0, 0, 0, 0, 0}
			return false
		}
		var hostType AddrType
		host, _port, hostType = dismantle(e.Host)
		reply = []byte{5, byte(REPSUCCESS), 0, byte(hostType)}
		switch hostType {
		case AddrV4, AddrV6:
			reply = append(reply, []byte(net.ParseIP(host))...) // host
			var portBytes []byte
			binary.BigEndian.PutUint16(portBytes, uint16(_port))
			reply = append(reply, portBytes...) // port
			return true
		default:
			reply = []byte{5, byte(REPERROR), 0, 0, 0, 0, 0, 0, 0, 0}
			return false
		}
	})
	return nil
}

func dismantle(host string) (hostOnly string, port int, hostType AddrType) {
	ipe := strings.Split(host, ":")
	_port := ipe[len(ipe)-1]
	port16, err := strconv.ParseInt(_port, 10, 16)
	if err != nil {
		panic("error port: " + _port)
	}
	port = int(port16)
	hostOnly = strings.TrimSuffix(host, ":"+_port)
	hostType = netAddrType2SocksAddrType(mynet.TypeOfAddr(host))
	return
}
