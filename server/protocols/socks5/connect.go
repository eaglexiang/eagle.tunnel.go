/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-02-24 17:56:31
 * @LastEditTime: 2019-08-28 19:51:07
 */

package socks5

import (
	"strconv"

	mynet "github.com/eaglexiang/go/net"
)

type connect struct {
}

func (conn connect) Handle(req []byte, e *mynet.Arg) (err error) {
	host, _port, err := getHostAndPort(req)
	if err != nil {
		return
	}
	port := strconv.FormatInt(int64(_port), 10)
	e.Host = host + ":" + port
	e.TheType = int(mynet.CONNECT)
	reply := "\u0005\u0000\u0000\u0001\u0000\u0000\u0000\u0000\u0000\u0000"

	err = e.Tunnel.WriteLeftStr(reply)
	return
}
