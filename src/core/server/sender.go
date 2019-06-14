/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-04 14:46:10
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-06-14 23:13:39
 */

package server

import (
	mynet "github.com/eaglexiang/go-net"
)

// Sender 请求发送者
type Sender interface {
	Send(e *mynet.Arg) error
	Name() string
}

// DefaultSender 注册sender的标准位置
var DefaultSender Sender
