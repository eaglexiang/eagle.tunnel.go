package config

import "testing"

func Test_getIPPort(t *testing.T) {
	ipe := "[9a:b2:a4:06:9a:81]"
	ip, port, err := getIPPort(ipe)
	if err != nil {
		t.Error(err)
	}
	if ip != "[9a:b2:a4:06:9a:81]" || port != "" {
		t.Error("[9a:b2:a4:06:9a:81] is ipv6")
	}

	ipe = "[9a:b2:a4:06:9a:81]:8080"
	ip, port, err = getIPPort(ipe)
	if err != nil {
		t.Error(err)
	}
	if ip != "[9a:b2:a4:06:9a:81]" || port != "8080" {
		t.Error("[9a:b2:a4:06:9a:81]:8080 is ipv6")
	}

	ipe = "[9a:b2:a4:06:9a:81"
	ip, port, err = getIPPort(ipe)
	if err == nil {
		t.Error("[9a:b2:a4:06:9a:81 is not ipv6")
	}

	ipe = "0.0.0.0"
	ip, port, err = getIPPort(ipe)
	if err != nil {
		t.Error(err)
	}
	if ip != "0.0.0.0" {
		t.Error("ip is not 0.0.0.0 but ", ip)
	}
}
