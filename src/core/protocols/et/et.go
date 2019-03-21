/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-27 08:24:57
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-03-19 20:08:58
 */

package et

import (
	"errors"

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
	subHandlers map[string]handler
	subSenders  map[int]sender
}

// CreateET 构造ET
func CreateET(arg *Arg) *ET {
	et := ET{
		arg:         arg,
		subHandlers: make(map[string]handler),
		subSenders:  make(map[int]sender),
	}
	dns := dNS{arg: arg, dnsResolver: mynet.ResolvIPv4, dnsType: EtDNS}
	dns6 := dNS{arg: arg, dnsResolver: mynet.ResolvIPv6, dnsType: EtDNS6}
	tcp := tCP{
		arg:  arg,
		dns:  dns,
		dns6: dns6,
	}
	location := location{arg: arg}
	check := check{arg: arg}

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
func (et *ET) AddSubHandler(h handler) {
	et.subHandlers[h.Name()] = h
}

// AddSubSender 添加子协议Sender
func (et *ET) AddSubSender(s sender) {
	et.subSenders[s.Type()] = s
}

// Match 判断请求消息是否匹配该业务
func (et *ET) Match(firstMsg []byte) bool {
	firstMsgStr := string(firstMsg)
	return firstMsgStr == et.arg.Head
}

func (et *ET) getHandler(subReq string) (handler, error) {
	h, ok := et.subHandlers[subReq]
	if !ok {
		return nil, errors.New("handler not found")
	}
	return h, nil
}

// Name Sender的名字
func (et *ET) Name() string {
	return "ET"
}
