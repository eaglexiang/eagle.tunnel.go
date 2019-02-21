/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-03 15:27:00
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-02-22 01:28:41
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
	firstMsg := getFirsMsg(conn) // 获取握手消息
	if firstMsg == nil {
		return errors.New("Relayer.Handle -> no firstMsg")
	}
	handler := getHandler(firstMsg, relayer.handlers) // 识别业务协议
	if handler == nil {
		ip := conn.RemoteAddr().String()
		return errors.New("Relayer.Handle -> no matched handler from " +
			ip + ": " +
			string(firstMsg))
	}

	// 进入业务流程
	tunnel := mytunnel.GetTunnel()
	defer mytunnel.PutTunnel(tunnel)
	tunnel.Left = &conn
	e := mynet.Arg{
		Msg:    firstMsg,
		Tunnel: tunnel,
	}

	err = handler.Handle(&e)
	if err != nil {
		if err.Error() == "no need to continue" {
			return nil
		}
		return errors.New("Relayer.Handle -> " +
			err.Error())
	}

	// 判断是否是sender业务
	typeOfSender := reflect.TypeOf(relayer.sender)
	typeOfHandler := reflect.TypeOf(handler)
	if typeOfHandler == typeOfSender {
		// sender业务直接流动
		tunnel.Flow()
		return nil
	}
	// 从非sender业务获取目的Host
	// 然后根据目的Host建立连接
	err = relayer.sender.Send(&e)
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

func getHandler(firstMsg []byte, handlers []Handler) Handler {
	var handler Handler
	for _, h := range handlers {
		if h.Match(firstMsg) {
			handler = h
			break
		}
	}
	return handler
}

func getFirsMsg(conn net.Conn) []byte {
	buffer := bytebuffer.GetKBBuffer()
	defer bytebuffer.PutKBBuffer(buffer)
	var err error
	buffer.Length, err = conn.Read(buffer.Buf())
	if err != nil {
		return nil
	}
	return buffer.Cut()
}
