/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-03 15:27:00
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-03-19 20:48:32
 */

package core

import (
	"errors"
	"net"

	logger "github.com/eaglexiang/eagle.tunnel.go/src/logger"
	"github.com/eaglexiang/go-bytebuffer"
	mynet "github.com/eaglexiang/go-net"
	mytunnel "github.com/eaglexiang/go-tunnel"
)

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
	tunnel := mytunnel.GetTunnel()
	tunnel.Left = conn
	tunnel.Timeout = Timeout
	firstMsg, handler, err := relay.shake(tunnel)
	if err != nil {
		return
	}
	defer bytebuffer.PutKBBuffer(firstMsg)
	e := &mynet.Arg{
		Msg:    firstMsg.Data(),
		Tunnel: tunnel,
	}
	relay.handleReqs(handler, tunnel, e)
}

func (relay *Relay) handleReqs(handler Handler,
	tunnel *mytunnel.Tunnel,
	e *mynet.Arg) {
	// 判断是否是sender业务
	var need2Continue bool
	if relay.sender.Name() == handler.Name() {
		need2Continue = relay.handleSenderReqs(handler, e)
	} else {
		need2Continue = relay.handleOtherReqs(handler, tunnel, e)
	}
	if need2Continue {
		tunnel.Flow()
	}
	mytunnel.PutTunnel(tunnel)
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
	tunnel *mytunnel.Tunnel,
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
		logger.Warning("fail to connect")
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
func (relay *Relay) shake(tunnel *mytunnel.Tunnel) (
	msg *bytebuffer.ByteBuffer,
	handler Handler, err error) {
	buffer := bytebuffer.GetKBBuffer()
	buffer.Length, err = tunnel.ReadLeft(buffer.Buf())
	if err != nil {
		bytebuffer.PutKBBuffer(buffer)
		logger.Warning("fail to get first msg")
		return nil, nil, err
	}
	handler, err = getHandler(buffer, relay.handlers)
	if err != nil {
		bytebuffer.PutKBBuffer(buffer)
		logger.Warning(err, ": ", buffer.String())
		return nil, nil, err
	}
	return buffer, handler, nil
}
