package eagletunnel

import (
	"errors"
	"net"
	"strconv"
	"strings"
	"time"

	"../eaglelib"
)

// ETTCP ET-TCP子协议的实现
type ETTCP struct {
}

// Send 发送请求
func (et *ETTCP) Send(e *NetArg) bool {
	var ok bool
	switch ProxyStatus {
	case ProxySMART:
		var inside bool
		el := ETLocation{}
		ok = el.Send(e)
		if ok {
			inside = e.boolObj
		} else {
			inside = false
		}
		if inside {
			ok = et.sendTCPReq2Server(e) == nil
		} else {
			ok = et.sendTCPReq2Remote(e) == nil
		}
	case ProxyENABLE:
		ok = et.sendTCPReq2Remote(e) == nil
	}
	return ok

}

func (et *ETTCP) sendTCPReq2Remote(e *NetArg) error {
	err := connect2Relayer(e.tunnel)
	if err != nil {
		return err
	}
	req := FormatEtType(EtTCP) + " " + e.ip + " " + strconv.Itoa(e.port)
	count, err := e.tunnel.WriteRight([]byte(req))
	if err != nil {
		return err
	}
	buffer := make([]byte, 1024)
	count, err = e.tunnel.ReadRight(buffer)
	if err != nil {
		return err
	}
	reply := string(buffer[:count])
	if reply != "ok" {
		err = errors.New("failed 2 connect 2 server by relayer")
	}
	return err
}

func (et *ETTCP) sendTCPReq2Server(e *NetArg) error {
	ipe := e.ip + ":" + strconv.Itoa(e.port)
	conn, err := net.DialTimeout("tcp", ipe, 5*time.Second)
	if err != nil {
		return err
	}
	e.tunnel.Right = &conn
	e.tunnel.EncryptRight = false
	return err
}

// Handle 处理ET-TCP请求
func (et *ETTCP) Handle(req Request, tunnel *eaglelib.Tunnel) bool {
	reqs := strings.Split(req.RequestMsgStr, " ")
	if len(reqs) < 3 {
		return false
	}
	ip := reqs[1]
	_port := reqs[2]
	port, err := strconv.ParseInt(_port, 10, 32)
	if err != nil {
		return false
	}
	e := NetArg{ip: ip, port: int(port), tunnel: tunnel}
	err = et.sendTCPReq2Server(&e)
	if err != nil {
		tunnel.WriteLeft([]byte("nok"))
		return false

	}
	_, err = tunnel.WriteLeft([]byte("ok"))
	return err == nil
}
