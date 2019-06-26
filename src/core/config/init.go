/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-27 08:37:36
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-06-26 21:42:40
 */

package config

import (
	"github.com/eaglexiang/eagle.tunnel.go/src/core/config/ipe"
	"github.com/eaglexiang/eagle.tunnel.go/src/core/protocols/et/comm"
	"github.com/eaglexiang/go-bytebuffer"
	"github.com/eaglexiang/go-logger"
	settings "github.com/eaglexiang/go-settings"
	"path"
	"time"
)

func init() {
	// 绑定不同的参数名
	settings.Bind("relayer", "relay")
	settings.Bind("channel", "data-key")
	// 设定参数默认值
	settings.SetDefault("timeout", "30")
	settings.SetDefault("location", "1;CN;CHN;China")
	settings.SetDefault("ip-type", "46")
	settings.SetDefault("data-key", "34")
	settings.SetDefault("head", "eagle_tunnel")
	settings.SetDefault("proxy-status", "smart")
	settings.SetDefault("user", "null:null")
	settings.SetDefault("user-check", "off")
	settings.SetDefault("listen", "0.0.0.0")
	settings.SetDefault("relay", "127.0.0.1")
	settings.SetDefault("http", "off")
	settings.SetDefault("socks", "off")
	settings.SetDefault("et", "off")
	settings.SetDefault("debug", "warning")
	settings.SetDefault("cipher", "simple")
	settings.SetDefault("maxclients", "0")
	settings.SetDefault("buffer.size", "1000")
	settings.SetDefault("dynamic-ipe", "off")
}

// initListens ipes的示例：192.168.0.1:8080,192.168.0.1:8081
func initListens() {
	ListenIPEs = ipe.ParseIPPortsSlice(settings.Get("listen"))
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
