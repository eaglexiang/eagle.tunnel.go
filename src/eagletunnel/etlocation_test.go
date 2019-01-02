/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-26 10:29:43
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-02 16:36:45
 */

package eagletunnel

import (
	"testing"
)

func Test_ETLocation_Send(t *testing.T) {
	InitConfig("../../tmp/t.conf")
	ExecConfig()
	testETLocationSend(t, "192.168.50.1", false)
	testETLocationSend(t, "127.0.0.1", false)
	testETLocationSend(t, "245.5.2.1", false)
	testETLocationSend(t, "221.2.55.32", false)
	testETLocationSend(t, "100.123.4.2", false)
	testETLocationSend(t, "43.45.102.33", true)
	testETLocationSend(t, "8.8.8.8", true)
}

func testETLocationSend(t *testing.T, ip string, proxy bool) {
	el := ETLocation{}
	e := NetArg{IP: ip}
	el.Send(&e)
	_proxy := CheckProxyByLocation(e.Reply)
	if _proxy != proxy {
		t.Error("代理情况错误 ", ip, " ", e.Reply)
	}
}
