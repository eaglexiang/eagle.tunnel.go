package config

import (
	"time"

	"github.com/eaglexiang/eagle.tunnel.go/src/core/config/ipe"
	myuser "github.com/eaglexiang/go-user"
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
