package et

import "testing"

func Test_Names(t *testing.T) {
	tcp := tCP{}
	testname(t, tcp, EtNameTCP)
	dns := dNS{dnsType: EtDNS}
	testname(t, dns, EtNameDNS)
	dns6 := dNS{dnsType: EtDNS6}
	testname(t, dns6, EtNameDNS6)
	l := location{}
	testname(t, l, EtNameLOCATION)
	ck := check{}
	testname(t, ck, EtNameCHECK)
	bd := bind{}
	testname(t, bd, EtNameBIND)
}

func Test_Types(t *testing.T) {
	tcp := tCP{}
	testtype(t, tcp, EtTCP)
	dns := dNS{dnsType: EtDNS}
	testtype(t, dns, EtDNS)
	dns6 := dNS{dnsType: EtDNS6}
	testtype(t, dns6, EtDNS6)
	l := location{}
	testtype(t, l, EtLOCATION)
	ck := check{}
	testtype(t, ck, EtCHECK)
	bd := bind{}
	testtype(t, bd, EtBIND)
}

func testname(t *testing.T, s handler, name string) {
	if s.Name() != name {
		t.Error("wrong name for ", name, ": ", s.Name())
	}
}

func testtype(t *testing.T, s handler, theType int) {
	if s.Type() != theType {
		t.Error("wrong type for ", FormatEtType(theType), ": ", FormatEtType(s.Type()))
	}
}
