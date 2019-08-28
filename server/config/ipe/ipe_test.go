/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-08-28 19:50:31
 * @LastEditTime: 2019-08-28 19:50:31
 */

package ipe

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

func Test_ToStrings(t *testing.T) {
	ipports := IPPorts{
		IP: "1.2.3.4",
		Ports: []string{
			"1000",
			"2000",
			"3000",
		},
	}

	ipPorts := ipports.ToStrings()
	if len(ipPorts) != 3 {
		t.Error(ipPorts)
	}
	if ipPorts[0] != "1.2.3.4:1000" {
		t.Error(ipPorts)
	}
	if ipPorts[1] != "1.2.3.4:2000" {
		t.Error(ipPorts)
	}
	if ipPorts[2] != "1.2.3.4:3000" {
		t.Error(ipPorts)
	}
}
