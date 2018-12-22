/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-13 19:04:31
 * @LastEditors: EagleXiang
 * @LastEditTime: 2018-12-22 13:32:01
 */

package eagletunnel

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"../eaglelib/src"
)

var ifProxyCache = CreateProxyCache() // [ip string, proxy bool]

// ETLocation ET-LOCATION子协议的实现
type ETLocation struct {
}

// Send 发送ET-LOCATION请求 解析IP是否适合直连。返回值表示是否成功解析，解析结果保存在e.boolObj
func (el *ETLocation) Send(e *NetArg) bool {
	if ifProxyCache.Exsit(e.IP) {
		proxy, err := ifProxyCache.Wait4Proxy(e.IP)
		if err != nil {
			return false
		}
		e.boolObj = proxy
		return true
	}
	if CheckPrivateIPv4(e.IP) {
		// 保留地址适合直连
		e.boolObj = true
		ifProxyCache.Add(e.IP)
		ifProxyCache.Update(e.IP, true)
		return true
	}
	err := el.checkInsideByRemote(e)
	if err != nil {
		e.boolObj = false
		return false
	}
	ifProxyCache.Update(e.IP, e.boolObj)
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
		if ifProxyCache.Exsit(ip) {
			proxy, err := ifProxyCache.Wait4Proxy(ip)
			if err != nil {
				reply = err.Error()
			}
			reply = strconv.FormatBool(proxy)
		} else {
			if CheckPrivateIPv4(ip) {
				reply = strconv.FormatBool(true)
				ifProxyCache.Add(ip)
				ifProxyCache.Update(ip, true)
			} else {
				proxy, err := CheckProxyByLocal(ip)
				if err != nil {
					reply = err.Error()
				} else {
					reply = strconv.FormatBool(proxy)
					ifProxyCache.Add(ip)
					ifProxyCache.Update(ip, proxy)
				}
			}
		}
		tunnel.WriteLeft([]byte(reply))
	}
}

// CheckProxyByLocal 本地解析IP是否需要使用代理
func CheckProxyByLocal(ip string) (bool, error) {
	location, err := CheckLocationByWeb(ip)
	if err != nil {
		return false, err
	}
	switch location {
	case "0;;;WRONG INPUT":
		err = errors.New("0;;;WRONG INPUT")
		return false, err
	case "1;ZZ;ZZZ;Reserved", "1;CN;CHN;China":
		return true, nil
	default:
		return false, nil
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
