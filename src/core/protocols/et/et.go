/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-27 08:24:57
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-03-19 20:08:58
 */

package et

import (
	mynet "github.com/eaglexiang/go-net"
	version "github.com/eaglexiang/go-version"
)

// ProtocolVersion 作为Sender使用的协议版本号
var ProtocolVersion, _ = version.CreateVersion("1.5")

// ProtocolCompatibleVersion 作为Handler可兼容的最低协议版本号
var ProtocolCompatibleVersion, _ = version.CreateVersion("1.3")

// ET ET代理协议的实现
// 必须使用CreateET来构造该结构
type ET struct {
	arg         *Arg
	subHandlers []Handler
	subSenders  map[int]Sender
}

// CreateET 构造ET
func CreateET(arg *Arg) *ET {
	et := ET{
		arg:        arg,
		subSenders: make(map[int]Sender),
	}
	dns := DNS{arg: arg, dnsResolver: mynet.ResolvIPv4}
	dns6 := DNS{arg: arg, dnsResolver: mynet.ResolvIPv6}
	tcp := TCP{
		arg:  arg,
		dns:  dns,
		dns6: dns6,
	}
	location := Location{arg: arg}
	check := Check{arg: arg}

	// 添加子协议的handler
	et.AddSubHandler(tcp)
	et.AddSubHandler(dns)
	et.AddSubHandler(dns6)
	et.AddSubHandler(location)
	et.AddSubHandler(check)

	// 添加子协议的sender
	et.AddSubSender(tcp)
	et.AddSubSender(dns)
	et.AddSubSender(dns6)
	et.AddSubSender(location)
	return &et
}

// AddSubHandler 添加ET子协议handler
func (et *ET) AddSubHandler(handler Handler) {
	et.subHandlers = append(et.subHandlers, handler)
}

// AddSubSender 添加子协议Sender
func (et *ET) AddSubSender(sender Sender) {
	et.subSenders[sender.Type()] = sender
}

// Match 判断请求消息是否匹配该业务
func (et *ET) Match(firstMsg []byte) bool {
	firstMsgStr := string(firstMsg)
	return firstMsgStr == et.arg.Head
}

func getHandler(subReq string, subHandlers []Handler) Handler {
	var handler Handler
	for _, h := range subHandlers {
		if h.Match(subReq) {
			handler = h
			break
		}
	}
	return handler
}

// Name Sender的名字
func (et *ET) Name() string {
	return "ET"
}
