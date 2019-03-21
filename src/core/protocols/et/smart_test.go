package et

import "testing"

func Test_ParseProxyStatus(t *testing.T) {
	ps, err := ParseProxyStatus("smart")
	if err != nil {
		t.Error(err)
	}
	if ps != ProxySMART {
		t.Error("wrong status: ", FormatProxyStatus(ps))
	}
	ps, err = ParseProxyStatus("smArt")
	if err != nil {
		t.Error(err)
	}
	if ps != ProxySMART {
		t.Error("wrong status: ", FormatProxyStatus(ps))
	}
	ps, err = ParseProxyStatus("SMART")
	if err != nil {
		t.Error(err)
	}
	if ps != ProxySMART {
		t.Error("wrong status: ", FormatProxyStatus(ps))
	}

	ps, err = ParseProxyStatus("enable")
	if err != nil {
		t.Error(err)
	}
	if ps != ProxyENABLE {
		t.Error("wrong status: ", FormatProxyStatus(ps))
	}
	ps, err = ParseProxyStatus("enabLe")
	if err != nil {
		t.Error(err)
	}
	if ps != ProxyENABLE {
		t.Error("wrong status: ", FormatProxyStatus(ps))
	}
	ps, err = ParseProxyStatus("ENABLE")
	if err != nil {
		t.Error(err)
	}
	if ps != ProxyENABLE {
		t.Error("wrong status: ", FormatProxyStatus(ps))
	}
}

func Test_FormatProxyStatus(t *testing.T) {
	ps := FormatProxyStatus(ProxyENABLE)
	if ps != ProxyEnableText {
		t.Error("wrong status for ", ProxyEnableText, ": ", ps)
	}
	ps = FormatProxyStatus(ProxySMART)
	if ps != ProxySmartText {
		t.Error("wrong status for ", ProxySmartText, ": ", ps)
	}
}
