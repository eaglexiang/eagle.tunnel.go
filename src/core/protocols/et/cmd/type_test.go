package cmd

import (
	"testing"

	"github.com/eaglexiang/eagle.tunnel.go/src/core/protocols/et/comm"
)

func Test_Names(t *testing.T) {
	tcp := TCP{}
	testname(t, tcp, comm.EtNameTCP)
	dns := DNS{DNSType: comm.EtDNS}
	testname(t, dns, comm.EtNameDNS)
	dns6 := DNS{DNSType: comm.EtDNS6}
	testname(t, dns6, comm.EtNameDNS6)
	l := Location{}
	testname(t, l, comm.EtNameLOCATION)
	ck := Check{}
	testname(t, ck, comm.EtNameCHECK)
	bd := Bind{}
	testname(t, bd, comm.EtNameBIND)
}

func Test_Types(t *testing.T) {
	tcp := TCP{}
	testtype(t, tcp, comm.EtTCP)
	dns := DNS{DNSType: comm.EtDNS}
	testtype(t, dns, comm.EtDNS)
	dns6 := DNS{DNSType: comm.EtDNS6}
	testtype(t, dns6, comm.EtDNS6)
	l := Location{}
	testtype(t, l, comm.EtLOCATION)
	ck := Check{}
	testtype(t, ck, comm.EtCHECK)
	bd := Bind{}
	testtype(t, bd, comm.EtBIND)
}

func testname(t *testing.T, s comm.Handler, name string) {
	if s.Name() != name {
		t.Error("wrong name for ", name, ": ", s.Name())
	}
}

func testtype(t *testing.T, s comm.Handler, theType int) {
	if s.Type() != theType {
		t.Error("wrong type for ",
			comm.FormatEtType(theType), ": ", comm.FormatEtType(s.Type()))
	}
}
