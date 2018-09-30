package eagletunnel

import (
	"testing"
)

func Test_ETLocation_Send(t *testing.T) {
	testETLocationSend(t, "192.168.50.1", true)
	testETLocationSend(t, "127.0.0.1", true)
	testETLocationSend(t, "245.5.2.1", true)
	testETLocationSend(t, "221.2.55.32", true)
	testETLocationSend(t, "100.123.4.2", true)
	testETLocationSend(t, "43.45.102.33", false)
	testETLocationSend(t, "8.8.8.8", false)
}

func testETLocationSend(t *testing.T, ip string, inside bool) {
	el := ETLocation{}
	e := NetArg{IP: ip}
	reuslt := el.Send(&e)
	if !reuslt {
		t.Error("解析失败")
	}
	if e.boolObj != inside {
		t.Error("本地地址应该直连")
	}
}
