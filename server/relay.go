/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-03 15:27:00
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-08-28 19:52:02
 */

package server

import (
	"errors"
	"net"

	"github.com/eaglexiang/eagle.tunnel.go/server/config"
	"github.com/eaglexiang/go/bytebuffer"
	"github.com/eaglexiang/go/logger"
	mynet "github.com/eaglexiang/go/net"
	"github.com/eaglexiang/go/tunnel"
)

// HTTPReqSize 允许HTTP请求的最大长度
const HTTPReqSize = 2e5

func init() {
	bytebuffer.RegisterPool(HTTPReqSize)
}

// Relay 网络入口，负责流量分发
// 必须使用CreateRelay方法进行构造
type Relay struct {
	handlers []Handler
	sender   Sender
}

// AddHandler 为Relayer增添可用的handler
func (relay *Relay) AddHandler(handler Handler) {
	relay.handlers = append(relay.handlers, handler)
}

// SetSender 为Relayer设置可用的Sender
func (relay *Relay) SetSender(sender Sender) {
	relay.sender = sender
}

// Handle 处理请求连接
func (relay *Relay) Handle(conn net.Conn) {
	var t = tunnel.GetTunnel()
	t.SetTimeout(config.Timeout)
	t.SetLeft(conn)
	firstMsg, handler, err := relay.shake(t)
	defer bytebuffer.PutBuffer(firstMsg)
	if err != nil {
		return
	}

	e := &mynet.Arg{
		Msg:    firstMsg.Cut(),
		Tunnel: t,
	}
	relay.handleReqs(handler, t, e)
}

func (relay *Relay) handleReqs(handler Handler,
	t *tunnel.Tunnel,
	e *mynet.Arg) {
	// 判断是否是sender业务
	var need2Continue bool
	if relay.sender.Name() == handler.Name() {
		need2Continue = relay.handleSenderReqs(handler, e)
	} else {
		need2Continue = relay.handleOtherReqs(handler, e)
	}
	if need2Continue {
		t.Flow()
	}
	tunnel.PutTunnel(t)
}

// 使用sender业务向远端发起请求
func (relay *Relay) handleSenderReqs(handler Handler,
	e *mynet.Arg) bool {
	err := handler.Handle(e)
	if err == nil {
		return true
	}
	if err.Error() != "no need to continue" {
		logger.Warning("Relay.Handle -> ", err)
	}
	return false
}

// 从非sender业务获取目的Host
// 然后根据目的Host建立连接
func (relay *Relay) handleOtherReqs(
	handler Handler,
	e *mynet.Arg) bool {
	// 获取Host
	err := handler.Handle(e)
	if err != nil {
		if err.Error() == "no need to continue" {
			return true
		}
		logger.Warning("fail to get host")
		return false
	}
	// 发起连接
	err = relay.sender.Send(e)
	if err != nil {
		logger.Warning("fail to connect ", e.Host)
		return false
	}
	// 完成委托行为
	for _, f := range e.Delegates {
		if !f() {
			return false
		}
	}
	return true
}

func getHandler(firstMsg *bytebuffer.ByteBuffer, handlers []Handler) (Handler, error) {
	for _, h := range handlers {
		if h.Match(firstMsg.Data()) {
			return h, nil
		}
	}
	return nil, errors.New("no matched handler")
}

// shake 握手
// 获取握手消息和对应handler
func (relay *Relay) shake(t *tunnel.Tunnel) (msg *bytebuffer.ByteBuffer, handler Handler, err error) {
	msg = bytebuffer.GetBuffer(HTTPReqSize)
	err = t.ReadLeft(msg)
	if err != nil {
		logger.Warning("fail to get first msg")
		return
	}
	handler, err = getHandler(msg, relay.handlers)
	if err != nil {
		logger.Warning(err, ": ", msg.String())
		return
	}
	return
}
