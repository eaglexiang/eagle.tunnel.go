/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-04 14:46:00
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-06-14 23:13:26
 */

package server

import mynet "github.com/eaglexiang/go-net"

// Handler 请求处理者
type Handler interface {
	Handle(e *mynet.Arg) error  // 处理业务
	Match(firstMsg []byte) bool // 判断业务请求是否符合该handler
	Name() string               // Handler的名字
}

// AllHandlers 注册handler的标准位置
var AllHandlers = make(map[string]Handler)
