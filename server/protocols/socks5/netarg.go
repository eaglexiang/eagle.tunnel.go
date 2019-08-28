/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-02-24 23:05:00
 * @LastEditTime: 2019-08-28 19:50:50
 */

package socks5

import (
	mynet "github.com/eaglexiang/go/net"
)

// AddrType 地址类型
type AddrType byte

// 网络地址类型
const (
	AddrERROR  AddrType = 0
	AddrV4              = 1
	AddrDomain          = 3
	AddrV6              = 4
)

func netAddrType2SocksAddrType(at mynet.AddrType) AddrType {
	switch at {
	case mynet.IPv4Addr:
		return AddrV4
	case mynet.DomainAddr:
		return AddrDomain
	case mynet.IPv6Addr:
		return AddrV6
	default:
		return AddrERROR
	}
}

func socksAddrType2NetAddrType(at AddrType) mynet.AddrType {
	switch at {
	case AddrV4:
		return mynet.IPv4Addr
	case AddrV6:
		return mynet.IPv6Addr
	case AddrDomain:
		return mynet.DomainAddr
	default:
		return mynet.InvalidAddr
	}
}

// NetOPType2SocksOPType 将net中的网络操作类型转换为socks5的网络操作类型
func NetOPType2SocksOPType(ot mynet.OpType) CMDType {
	switch ot {
	case mynet.BIND:
		return BIND
	case mynet.CONNECT:
		return CONNECT
	case mynet.UDP:
		return UDP
	default:
		return ERROR
	}
}

// SocksOPType2NetOPType 将socks5的网络操作类型转换为net中的网络操作类型
func SocksOPType2NetOPType(ct CMDType) mynet.OpType {
	switch ct {
	case BIND:
		return mynet.BIND
	case CONNECT:
		return mynet.CONNECT
	case UDP:
		return mynet.UDP
	default:
		return mynet.ERROR
	}
}
