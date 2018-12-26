/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-13 19:04:31
 * @LastEditors: EagleXiang
 * @LastEditTime: 2018-12-26 10:49:33
 */

package eagletunnel

import (
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"

	"../eaglelib/src"
)

// IfProxyCache 缓存IP是否需要被代理
var IfProxyCache = CreateProxyCache() // [ip string, proxy bool]

// ETLocation ET-LOCATION子协议的实现
type ETLocation struct {
}

// Send 发送ET-LOCATION请求 解析IP是否适合代理。返回值表示是否成功解析，解析结果保存在e.boolObj
func (el *ETLocation) Send(e *NetArg) bool {
	if IfProxyCache.Exsit(e.IP) {
		// 读取缓存
		proxy, err := IfProxyCache.Wait4Proxy(e.IP)
		if err != nil {
			e.boolObj = true
			return false
		}
		e.boolObj = proxy
		return true
	}
	if CheckPrivateIPv4(e.IP) {
		// 保留地址不适合代理
		e.boolObj = false
		IfProxyCache.Add(e.IP)
		IfProxyCache.Update(e.IP, e.boolObj)
		return true
	}
	err := el.checkProxyByRemote(e)
	if err != nil {
		// 解析失败，尝试直连
		println("fail to resolv location by remote: ", err.Error())
		e.boolObj = true
		return false
	}
	IfProxyCache.Update(e.IP, e.boolObj)
	return true
}

func (el *ETLocation) checkProxyByRemote(e *NetArg) error {
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
	reply := string(buffer[:count])
	e.boolObj, err = strconv.ParseBool(reply)
	return err
}

// Handle 处理ET-LOCATION请求
func (el *ETLocation) Handle(req Request, tunnel *eaglelib.Tunnel) {
	reqs := strings.Split(req.RequestMsgStr, " ")
	if len(reqs) < 2 {
		return
	}

	// read cache
	var reply string
	ip := reqs[1]
	if IfProxyCache.Exsit(ip) {
		proxy, err := IfProxyCache.Wait4Proxy(ip)
		if err != nil {
			reply = err.Error()
		}
		reply = strconv.FormatBool(proxy)
		tunnel.WriteLeft([]byte(reply))
		return
	}

	// check private
	if CheckPrivateIPv4(ip) {
		reply = strconv.FormatBool(false)
		IfProxyCache.Add(ip)
		IfProxyCache.Update(ip, false)
		tunnel.WriteLeft([]byte(reply))
		return
	}

	// check location
	proxy, err := CheckProxyByLocal(ip)
	if err != nil {
		reply = err.Error()
	} else {
		reply = strconv.FormatBool(proxy)
		IfProxyCache.Add(ip)
		IfProxyCache.Update(ip, proxy)
	}
	tunnel.WriteLeft([]byte(reply))
}

// CheckProxyByLocal 本地解析IP是否需要使用代理
func CheckProxyByLocal(ip string) (bool, error) {
	_ip := net.ParseIP(ip)
	if _ip.To4() == nil {
		println("ipv6: ", ip)
		return true, errors.New("dont suport ipv6")
	}
	location, err := CheckLocationByWeb(ip)
	if err != nil {
		return false, err
	}
	switch location {
	case "0;;;WRONG INPUT":
		err = errors.New("0;;;WRONG INPUT")
		return true, err
	case "1;ZZ;ZZZ;Reserved", "1;CN;CHN;China":
		return false, nil
	default:
		return true, nil
	}
}

// CheckLocationByWeb 外部解析IP的Location
func CheckLocationByWeb(ip string) (string, error) {
	req := "https://ip2c.org/" + ip
	res, err := http.Get(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	bodyStr := string(body)
	return bodyStr, nil
}
