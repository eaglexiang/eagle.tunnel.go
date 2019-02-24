/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-02-24 23:05:00
 * @LastEditTime: 2019-02-25 00:57:44
 */

package socks5

import (
	mynet "github.com/eaglexiang/go-net"
)

// 网络地址类型
const (
	AddrERROR  = 0
	AddrV4     = 1
	AddrDomain = 3
	AddrV6     = 4
)

func netAddrType2SocksAddrType(netAddrType int) int {
	switch netAddrType {
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

func socksAddrType2NetAddrType(socksAddrType int) int {
	switch socksAddrType {
	case AddrV4:
		return mynet.IPv4Addr
	case AddrV6:
		return mynet.IPv6Addr
	case AddrDomain:
		return mynet.DomainAddr
	default:
		return mynet.ErrorAddr
	}
}

// NetOPType2SocksOPType 将net中的网络操作类型转换为socks5的网络操作类型
func NetOPType2SocksOPType(netOPType int) int {
	switch netOPType {
	case mynet.BIND:
		return SOCKSBIND
	case mynet.CONNECT:
		return SOCKSCONNECT
	case mynet.UDP:
		return SOCKSUDP
	default:
		return SOCKSERROR
	}
}

// SocksOPType2NetOPType 将socks5的网络操作类型转换为net中的网络操作类型
func SocksOPType2NetOPType(socksOPType int) int {
	switch socksOPType {
	case SOCKSBIND:
		return mynet.BIND
	case SOCKSCONNECT:
		return mynet.CONNECT
	case SOCKSUDP:
		return mynet.UDP
	default:
		return mynet.Error
	}
}
