/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-08-24 10:50:13
 * @LastEditTime: 2019-09-08 12:02:18
 */
package et

import (
	"testing"

	"github.com/eaglexiang/eagle.tunnel.go/server/protocols/et/comm"
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
		RemoteIPE: comm.GetSetting("relay"),
		Head:      "testHead",
	}
	smartArg := comm.SmartArg{
		ProxyStatus:   comm.ProxyENABLE,
		LocalLocation: comm.GetSetting("location"),
	}

	return &comm.Arg{
		ConnArg:  connArg,
		SmartArg: smartArg,
		UsersArg: users,
		IPType:   comm.GetSetting("ip-type"),
	}
}
