/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-04-01 22:05:31
 * @LastEditTime: 2019-08-24 11:49:52
 */

package cmd

import (
	"testing"

	"github.com/eaglexiang/eagle.tunnel.go/server/protocols/et/comm"
)

func Test_Names(t *testing.T) {
	tcp := TCP{}
	testname(t, tcp, comm.TCPTxt)
	dns := DNS{DNSType: comm.DNS}
	testname(t, &dns, comm.DNSTxt)
	dns6 := DNS{DNSType: comm.DNS6}
	testname(t, &dns6, comm.DNS6Txt)
	l := Location{}
	testname(t, &l, comm.LOCATIONTxt)
	ck := Check{}
	testname(t, ck, comm.CHECKTxt)
	bd := Bind{}
	testname(t, bd, comm.BINDTxt)
}

func Test_Types(t *testing.T) {
	tcp := TCP{}
	testtype(t, tcp, comm.TCP)
	dns := DNS{DNSType: comm.DNS}
	testtype(t, &dns, comm.DNS)
	dns6 := DNS{DNSType: comm.DNS6}
	testtype(t, &dns6, comm.DNS6)
	l := Location{}
	testtype(t, &l, comm.LOCATION)
	ck := Check{}
	testtype(t, ck, comm.CHECK)
	bd := Bind{}
	testtype(t, bd, comm.BIND)
}

func testname(t *testing.T, s comm.Handler, name string) {
	if s.Name() != name {
		t.Error("wrong name for ", name, ": ", s.Name())
	}
}

func testtype(t *testing.T, s comm.Handler, theType comm.CMDType) {
	if s.Type() != theType {
		t.Error("wrong type for ",
			comm.FormatEtType(theType), ": ", comm.FormatEtType(s.Type()))
	}
}
