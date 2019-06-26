package config

import (
	"strings"

	"github.com/eaglexiang/eagle.tunnel.go/src/core/protocols/et/comm"
	"github.com/eaglexiang/go-settings"
)

// CreateETArg 构建ET.Arg
func CreateETArg(relay string) *comm.Arg {

	users := comm.UsersArg{
		LocalUser:  LocalUser,
		ValidUsers: Users,
	}
	connArg := comm.ConnArg{
		RemoteIPE: relay,
		Head:      settings.Get("head"),
	}
	smartArg := comm.SmartArg{
		ProxyStatus:   ProxyStatus,
		LocalLocation: settings.Get("location"),
	}

	return &comm.Arg{
		ConnArg:  connArg,
		SmartArg: smartArg,
		UsersArg: users,
		IPType:   settings.Get("ip-type"),
	}
}

func handleSingleHost(host string) {
	// 将所有连续空格缩减为单个空格
	for {
		newHost := strings.Replace(host, "  ", " ", -1)
		if newHost == host {
			break
		}
		host = newHost
	}

	if host == "" {
		return
	}

	items := strings.Split(host, " ")
	if len(items) < 2 {
		panic("invalid hosts line: " + host)
	}
	ip := strings.TrimSpace(items[0])
	domain := strings.TrimSpace(items[1])
	if domain != "" && ip != "" {
		comm.HostsCache[domain] = ip
	} else {
		panic("invalid hosts line: " + host)
	}
}
