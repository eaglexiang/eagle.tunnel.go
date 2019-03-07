/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-03 15:27:00
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-03-08 02:34:14
 */

package core

import (
	"errors"
	"net"
	"reflect"

	"github.com/eaglexiang/go-bytebuffer"
	mynet "github.com/eaglexiang/go-net"
	mytunnel "github.com/eaglexiang/go-tunnel"
)

// Relayer 网络入口，负责流量分发
// 必须使用CreateRelayer方法进行构造
type Relayer struct {
	handlers []Handler
	sender   Sender
}

// AddHandler 为Relayer增添可用的handler
func (relayer *Relayer) AddHandler(handler Handler) {
	relayer.handlers = append(relayer.handlers, handler)
}

// SetSender 为Relayer设置可用的Sender
func (relayer *Relayer) SetSender(sender Sender) {
	relayer.sender = sender
}

// Handle 处理请求连接
func (relayer *Relayer) Handle(conn net.Conn) (err error) {
	tunnel := mytunnel.GetTunnel()
	defer mytunnel.PutTunnel(tunnel)
	tunnel.Left = conn
	tunnel.Timeout = Timeout
	// 获取握手消息和对应handler
	firstMsg, handler, err := relayer.shake(tunnel)
	if err != nil {
		return errors.New("Relayer.Handle -> " +
			err.Error())
	}
	defer bytebuffer.PutKBBuffer(firstMsg)
	e := &mynet.Arg{
		Msg:    firstMsg.Data(),
		Tunnel: tunnel,
	}
	return relayer.handleReqs(handler, tunnel, e)
}

func (relayer *Relayer) handleReqs(handler Handler,
	tunnel *mytunnel.Tunnel,
	e *mynet.Arg) error {
	// 判断是否是sender业务
	if reflect.TypeOf(relayer.sender) == reflect.TypeOf(handler) {
		return relayer.handleSenderReqs(handler, tunnel, e)
	}
	return relayer.handleOtherReqs(handler, tunnel, e)
}

// 使用sender业务向远端发起请求
func (relayer *Relayer) handleSenderReqs(handler Handler,
	tunnel *mytunnel.Tunnel,
	e *mynet.Arg) (err error) {
	// 直接处理
	err = handler.Handle(e)
	if err != nil {
		if err.Error() == "no need to continue" {
			return nil
		}
		return errors.New("Relayer.Handle -> " +
			err.Error())
	}
	// 开始流动
	tunnel.Flow()
	return nil
}

// 从非sender业务获取目的Host
// 然后根据目的Host建立连接
func (relayer *Relayer) handleOtherReqs(
	handler Handler,
	tunnel *mytunnel.Tunnel,
	e *mynet.Arg) (err error) {
	// 获取Host
	err = handler.Handle(e)
	if err != nil {
		if err.Error() == "no need to continue" {
			return nil
		}
		return errors.New("Relayer.Handle -> " +
			err.Error())
	}
	// 发起连接
	err = relayer.sender.Send(e)
	if err != nil {
		return errors.New("Relayer.Handle -> " +
			err.Error())
	}
	// 完成委托行为
	for _, f := range e.Delegates {
		f()
	}
	tunnel.Flow()
	return nil
}

func getHandler(firstMsg *bytebuffer.ByteBuffer, handlers []Handler) Handler {
	var handler Handler
	for _, h := range handlers {
		if h.Match(firstMsg.Data()) {
			handler = h
			break
		}
	}
	return handler
}

func (relayer *Relayer) shake(tunnel *mytunnel.Tunnel) (
	msg *bytebuffer.ByteBuffer,
	handler Handler, err error) {
	buffer := bytebuffer.GetKBBuffer()
	buffer.Length, err = tunnel.ReadLeft(buffer.Buf())
	if err != nil {
		bytebuffer.PutKBBuffer(buffer)
		return nil, nil, errors.New("getFirstMsg -> " +
			err.Error())
	}
	handler = getHandler(buffer, relayer.handlers)
	if handler == nil {
		bytebuffer.PutKBBuffer(buffer)
		return nil, nil, errors.New("Relayer.Handle -> no matched handler from " +
			tunnel.Left.RemoteAddr().String() + ": ")
	}
	return buffer, handler, nil
}
