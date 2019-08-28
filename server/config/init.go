/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-27 08:37:36
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-08-28 20:47:38
 */

package config

import (
	"path"
	"time"

	"github.com/eaglexiang/eagle.tunnel.go/server/config/ipe"
	"github.com/eaglexiang/eagle.tunnel.go/server/protocols/et/comm"
	"github.com/eaglexiang/go/bytebuffer"
	"github.com/eaglexiang/go/logger"
	"github.com/eaglexiang/go/settings"
)

func init() {
	// 绑定不同的参数名
	settings.Bind("channel", "data-key")
	// 设定参数默认值
	settings.SetDefault("timeout", "30")
	settings.SetDefault("data-key", "34")
	settings.SetDefault("head", "eagle_tunnel")
	settings.SetDefault("proxy-status", "enable")
	settings.SetDefault("user", "null:null")
	settings.SetDefault("user-check", "off")
	settings.SetDefault("listen", "0.0.0.0")
	settings.SetDefault("http", "off")
	settings.SetDefault("socks", "off")
	settings.SetDefault("et", "off")
	settings.SetDefault("debug", "warning")
	settings.SetDefault("cipher", "simple")
	settings.SetDefault("maxclients", "0")
	settings.SetDefault("buffer.size", "10000")
	settings.SetDefault("dynamic-ipe", "0")
}

// initListens ipes的示例：192.168.0.1:8080,192.168.0.1:8081
func initListens() {
	ListenIPEs = ipe.ParseIPPortsSlice(settings.Get("listen"))

	randPortsCount := settings.GetInt64("dynamic-ipe")
	for _, ipe := range ListenIPEs {
		ipe.RandPorts(int(randPortsCount))
	}
}

func initRelays() {
	RelayIPEs = ipe.ParseIPPortsSlice(settings.Get("relay"))
}

func RelayIPE() string {
	relayIPPorts := RelayIPEs[0]
	relayIPE := relayIPPorts.IP + ":" + relayIPPorts.Ports[0]
	return relayIPE
}

func initLogger() {
	logger.SetGrade(settings.Get("debug"))
}

func initTimeout() {
	timeout := settings.GetInt64("timeout")
	Timeout = time.Second * time.Duration(timeout)
	comm.Timeout = Timeout
}

func initBufferSize() {
	size := settings.GetInt64("buffer.size")
	bytebuffer.SetDefaultSize(int(size))
}

func initLocalUser() {
	// 读取本地用户
	if !settings.Exsit("user") {
		SetUser("null:null")
	} else {
		SetUser(settings.Get("user"))
	}
}

// initUserList 初始化用户列表
func initUserList() {
	if settings.Get("user-check") != "on" {
		return
	}

	usersPath := path.Join(settings.Get("config-dir"), "/users.list")
	importUsers(usersPath)
}

func initProxyStatus() {
	var err error
	s := settings.Get("proxy-status")
	ProxyStatus, err = comm.ParseProxyStatus(s)
	if err != nil {
		panic(err)
	}
}
