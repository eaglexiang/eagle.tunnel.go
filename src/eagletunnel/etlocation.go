/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-13 19:04:31
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-03 01:44:01
 */

package eagletunnel

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"../eaglelib/src"
)

// IPGeoCacheClient Client持有的IP-Geo数据缓存
var IPGeoCacheClient = CreateLocationCache() // [ip string, location string]
// IPGeoCacheServer Server持有的IP-Geo数据缓存
var IPGeoCacheServer = CreateLocationCache() // [ip string, location string]

// ETLocation ET-LOCATION子协议的实现
type ETLocation struct {
}

// Send 发送ET-LOCATION请求 解析IP的地理位置，结果存放于e.Reply
func (el *ETLocation) Send(e *NetArg) {
	if IPGeoCacheClient.Exsit(e.IP) {
		// 读取缓存
		location, err := IPGeoCacheClient.Wait4Proxy(e.IP)
		if err != nil {
			e.Reply = err.Error()
			return
		}
		e.Reply = location
		return
	}

	IPGeoCacheClient.Add(e.IP)
	if CheckPrivateIPv4(e.IP) {
		// 保留地址不适合代理
		e.Reply = "1;ZZ;ZZZ;Reserved"
		IPGeoCacheClient.Update(e.IP, e.Reply)
		return
	}
	err := el.checkLocationByRemote(e)
	if err != nil {
		// 解析失败，尝试直连
		println("fail to resolv location by remote: ", err.Error())
		e.Reply = "0;;;WRONG INPUT"
		IPGeoCacheClient.Delete(e.IP)
		return
	}
	IPGeoCacheClient.Update(e.IP, e.Reply)
}

func (el *ETLocation) checkLocationByRemote(e *NetArg) error {
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
	e.Reply = string(buffer[:count])
	return nil
}

// Handle 处理ET-LOCATION请求
func (el *ETLocation) Handle(req Request, tunnel *eaglelib.Tunnel) {
	reqs := strings.Split(req.RequestMsgStr, " ")
	if len(reqs) < 2 {
		return
	}

	// read cache
	ip := reqs[1]
	if IPGeoCacheServer.Exsit(ip) {
		location, err := IPGeoCacheServer.Wait4Proxy(ip)
		if err != nil {
			tunnel.WriteLeft([]byte(err.Error()))
		} else {
			tunnel.WriteLeft([]byte(location))
		}
		return
	}

	IPGeoCacheServer.Add(ip)

	// check location
	location, err := CheckLocationByWeb(ip)
	if err != nil {
		fmt.Println("fail to resolv location by web: " + err.Error())
		IPGeoCacheServer.Delete(ip)
		tunnel.WriteLeft([]byte(err.Error()))
	} else {
		IPGeoCacheServer.Update(ip, location)
		tunnel.WriteLeft([]byte(location))
	}

}

// CheckProxyByLocation 本地解析IP是否需要使用代理
func CheckProxyByLocation(location string) bool {
	switch location {
	case "0;;;WRONG INPUT":
		return true
	case "1;ZZ;ZZZ;Reserved", ConfigKeyValues["location"]:
		return false
	default:
		return true
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
