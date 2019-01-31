/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-04 14:46:10
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-31 21:09:45
 */

package etcore

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
