/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-02-24 17:56:31
 * @LastEditTime: 2019-02-24 18:19:21
 */

package socks5

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"

	mynet "github.com/eaglexiang/go-net"
)

type connect struct {
}

func (conn connect) Handle(req []byte, e *mynet.Arg) error {
	ip, err := conn.getHost(req)
	if err != nil {
		return errors.New("connect.Handle -> " + err.Error())
	}
	_port := conn.getPort(req)
	port := strconv.FormatInt(int64(_port), 10)
	if _port <= 0 {
		return errors.New("connect.Handle -> invalid des: " +
			ip + ":" + port)
	}
	e.Host = ip + ":" + port
	reply := "\u0005\u0000\u0000\u0001\u0000\u0000\u0000\u0000\u0000\u0000"
	_, err = e.Tunnel.WriteLeft([]byte(reply))
	if err != nil {
		return errors.New("connect.Handle -> " + err.Error())
	}
	return nil
}

func (conn connect) getHost(request []byte) (string, error) {
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
		return "", errors.New("connect.getHost -> invalid socks req des type: " +
			strconv.FormatInt(int64(destype), 10))
	}
}

func (conn connect) getPort(request []byte) int {
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
		err = errors.New("connect.getPort -> invalid destype")
	}
	if err == nil {
		port = int16(binary.BigEndian.Uint16(buffer))
	}
	return int(port)
}
