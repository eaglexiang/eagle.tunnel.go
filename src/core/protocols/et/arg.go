/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-02-21 18:37:43
 * @LastEditTime: 2019-02-21 23:08:25
 */

package et

import (
	"time"

	myuser "github.com/eaglexiang/go-user"
)

// Arg 启动ET协议需要的的参数集
// 此参数集用于启动和配置ET协议服务
type Arg struct {
	ConnArg
	UsersArg
	SmartArg
	IPType string `label:"4/6/46/64 采用什么DNS解析模式"`
}

// UsersArg 用户系统使用的参数集
type UsersArg struct {
	LocalUser  *myuser.User
	ValidUsers map[string]*myuser.User
}

// ConnArg 连接相关参数集
type ConnArg struct {
	RemoteIPE string // IP:Port
	Timeout   time.Duration
	Head      string // 协议头
}

// SmartArg 智能模式需要的参数集
type SmartArg struct {
	ProxyStatus   int
	LocalLocation string
}
