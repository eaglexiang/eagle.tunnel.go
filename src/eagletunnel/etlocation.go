package eagletunnel

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/eaglexiang/eagle.lib.go/src"
)

// ETLocation ET-LOCATION子协议的实现
type ETLocation struct {
}

// Send 发送请求
func (el *ETLocation) Send(e *NetArg) bool {
	var err error
	_inside, ok := insideCache.Load(e.IP)
	if ok {
		e.boolObj, _ = _inside.(bool)
	} else {
		err = el.checkInsideByRemote(e)
		if err == nil {
			insideCache.Store(e.IP, e.boolObj)
		} else {
			var inside bool
			inside, err = CheckInsideByLocal(e.IP)
			if err == nil {
				e.boolObj = inside
				insideCache.Store(e.IP, e.boolObj)
			}
		}
	}
	return err == nil
}

func (el *ETLocation) checkInsideByRemote(e *NetArg) error {
	tunnel := eaglelib.Tunnel{}
	defer tunnel.Close()
	err := connect2Relayer(&tunnel)
	if err != nil {
		return err
	}
	req := FormatEtType(EtLOCATION) + " " + e.IP
	var count int
	count, err = tunnel.WriteRight([]byte(req))
	if err != nil {
		return err
	}
	buffer := make([]byte, 1024)
	count, err = tunnel.ReadRight(buffer)
	if err != nil {
		return err
	}
	e.boolObj, err = strconv.ParseBool(string(buffer[0:count]))
	return err
}

// Handle 处理ET-LOCATION请求
func (el *ETLocation) Handle(req Request, tunnel *eaglelib.Tunnel) {
	reqs := strings.Split(req.RequestMsgStr, " ")
	if len(reqs) >= 2 {
		var reply string
		ip := reqs[1]
		_inside, ok := insideCache.Load(ip)
		if ok {
			inside := _inside.(bool)
			reply = strconv.FormatBool(inside)
		} else {
			var err error
			inside, err := CheckInsideByLocal(ip)
			if err != nil {
				reply = fmt.Sprint(err)
			} else {
				reply = strconv.FormatBool(inside)
				insideCache.Store(ip, inside)
			}
		}
		tunnel.WriteLeft([]byte(reply))
	}
}
