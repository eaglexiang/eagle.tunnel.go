/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-02-24 17:56:31
 * @LastEditTime: 2019-02-24 18:46:16
 */

package socks5

import (
	"errors"
	"strconv"

	mynet "github.com/eaglexiang/go-net"
)

type connect struct {
}

func (conn connect) Handle(req []byte, e *mynet.Arg) error {
	ip, err := getHost(req)
	if err != nil {
		return errors.New("connect.Handle -> " + err.Error())
	}
	_port, err := getPort(req)
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
