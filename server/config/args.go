/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-08-24 10:43:53
 * @LastEditTime: 2019-08-28 19:48:59
 */

package config

import (
	"time"

	"github.com/eaglexiang/eagle.tunnel.go/server/config/ipe"
	myuser "github.com/eaglexiang/go/user"
)

// LocalUser 本地用户
var (
	LocalUser *myuser.ValidUser
	// Users 所有授权用户
	Users map[string]*myuser.ValidUser
	// ConfigPath 主配置文件的路径
	ConfigPath string
	// ProxyStatus 代理的状态（全局/智能）
	ProxyStatus int
	// Timeout 超时时间
	Timeout    time.Duration
	ListenIPEs []*ipe.IPPorts
	RelayIPEs  []*ipe.IPPorts
)
