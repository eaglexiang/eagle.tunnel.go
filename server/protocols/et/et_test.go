/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-08-24 10:50:13
 * @LastEditTime: 2019-08-28 20:29:40
 */
package et

import (
	"testing"

	"github.com/eaglexiang/eagle.tunnel.go/server/protocols/et/comm"
	"github.com/eaglexiang/go/settings"
)

func Test_ET(t *testing.T) {
	e := createETArg()
	et := NewET(e)
	if !et.Match([]byte("testHead")) {
		t.Error("testHead doesn't work")
	}
}

func createETArg() *comm.Arg {
	users := comm.UsersArg{
		LocalUser:  nil,
		ValidUsers: nil,
	}
	connArg := comm.ConnArg{
		RemoteIPE: settings.Get("relay"),
		Head:      "testHead",
	}
	smartArg := comm.SmartArg{
		ProxyStatus:   comm.ProxyENABLE,
		LocalLocation: settings.Get("location"),
	}

	return &comm.Arg{
		ConnArg:  connArg,
		SmartArg: smartArg,
		UsersArg: users,
		IPType:   settings.Get("ip-type"),
	}
}
