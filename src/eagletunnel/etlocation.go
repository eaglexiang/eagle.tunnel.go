package eagletunnel

import (
	"strconv"
	"strings"

	"github.com/eaglexiang/eagle.lib.go/src"
)

// ETLocation ET-LOCATION子协议的实现
type ETLocation struct {
}

// Send 发送ET-LOCATION请求 解析IP是否适合直连。返回值表示是否成功解析，解析结果保存在e.boolObj
func (el *ETLocation) Send(e *NetArg) bool {
	_inside, ok := insideCache.Load(e.IP)
	if ok {
		e.boolObj, _ = _inside.(bool)
		return true
	}
	if CheckPrivateIPv4(e.IP) {
		// 保留地址适合直连
		e.boolObj = true
		insideCache.Store(e.IP, true)
		return true
	}
	err := el.checkInsideByRemote(e)
	if err != nil {
		// 远端不响应，不得已本地解析。尽量使用远端解析可减少外部API的负载
		var inside bool
		inside, err = CheckInsideByLocal(e.IP)
		if err != nil {
			return false
		}
		e.boolObj = inside
		insideCache.Store(e.IP, e.boolObj)
		return true
	}
	insideCache.Store(e.IP, e.boolObj)
	return true
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
			if CheckPrivateIPv4(ip) {
				reply = strconv.FormatBool(true)
				insideCache.Store(ip, true)
			} else {
				var err error
				inside, err := CheckInsideByLocal(ip)
				if err != nil {
					reply = err.Error()
				} else {
					reply = strconv.FormatBool(inside)
					insideCache.Store(ip, inside)
				}
			}
		}
		tunnel.WriteLeft([]byte(reply))
	}
}
