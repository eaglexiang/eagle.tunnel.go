/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-27 08:24:57
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-08-24 11:48:57
 */

package et

import (
	"net"

	"github.com/eaglexiang/go-tunnel"

	"github.com/eaglexiang/eagle.tunnel.go/server/protocols/et/cmd"
	"github.com/eaglexiang/eagle.tunnel.go/server/protocols/et/comm"
	"github.com/eaglexiang/go-logger"
	mynet "github.com/eaglexiang/go-net"
)

// ET ET代理协议的实现
// 必须使用NewET进行初始化
type ET struct{}

// NewET 构造ET
func NewET(arg *comm.Arg) *ET {
	comm.DefaultArg = arg

	et := ET{}
	dns := cmd.DNS{DNSResolver: mynet.ResolvIPv4, DNSType: comm.DNS}
	dns6 := cmd.DNS{DNSResolver: mynet.ResolvIPv6, DNSType: comm.DNS6}
	tcp := cmd.TCP{}
	location := cmd.Location{}
	check := cmd.NewCheck()

	// 添加子协议的handler
	comm.AddSubHandler(tcp)
	comm.AddSubHandler(&dns)
	comm.AddSubHandler(&dns6)
	comm.AddSubHandler(&location)
	comm.AddSubHandler(check)

	// 添加子协议的sender
	comm.AddSubSender(tcp)
	comm.AddSubSender(&dns)
	comm.AddSubSender(&dns6)
	comm.AddSubSender(&location)

	comm.Connect2Remote = et.connect2Relay

	return &et
}

// Match 判断请求消息是否匹配该业务
func (et *ET) Match(firstMsg []byte) bool {
	firstMsgStr := string(firstMsg)
	return firstMsgStr == comm.DefaultArg.Head
}

// Name Sender的名字
func (et *ET) Name() string {
	return "ET"
}

// connect2Relay 连接到下一个Relay，完成版本校验和用户校验两个步骤
func (et *ET) connect2Relay(t *tunnel.Tunnel) error {
	conn, err := net.DialTimeout("tcp", comm.DefaultArg.RemoteIPE, comm.Timeout)
	if err != nil {
		logger.Warning(err)
		return err
	}
	t.SetRight(conn)
	err = et.checkVersionOfRelayer(t)
	if err != nil {
		return err
	}
	return et.checkLocalUser(t)
}
