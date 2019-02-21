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
	ProxyStatus   int
	IPType        string   // 4/6/46/64
	Head          string   // 协议头
	RemoteET      string   // IP:Port
	LocalLocation string   // 供智能模式用来判断是否需要代理
	Users         UsersArg // 用户系统使用的参数集
	Timeout       time.Duration
}

// UsersArg 用户系统使用的参数集
type UsersArg struct {
	LocalUser  *myuser.User
	ValidUsers map[string]*myuser.User
}
