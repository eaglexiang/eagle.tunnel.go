/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-13 19:04:31
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-03-03 05:40:42
 */

package et

import (
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	mynet "github.com/eaglexiang/go-net"
	mytunnel "github.com/eaglexiang/go-tunnel"
)

// iPGeoCacheClient Client持有的IP-Geo数据缓存
var iPGeoCacheClient = CreateLocationCache() // [ip string, location string]
// iPGeoCacheServer Server持有的IP-Geo数据缓存
var iPGeoCacheServer = CreateLocationCache() // [ip string, location string]

// Location ET-LOCATION子协议的实现
type Location struct {
	arg *Arg
}

// Send 发送ET-LOCATION请求 解析IP的地理位置，结果存放于e.Reply
func (l Location) Send(et *ET, e *NetArg) (err error) {
	node, loaded := iPGeoCacheClient.Get(e.IP)
	if loaded {
		e.Location, err = node.Wait()
		if err != nil {
			err = errors.New("Location.Send -> " + err.Error())
		}
		return
	}
	ip := net.ParseIP(e.IP)
	if ip == nil {
		return errors.New("Location.Send -> invalid ip")
	}
	if ip.To4() == nil {
		// IPv6 默认代理
		e.Location = "Ipv6"
		iPGeoCacheClient.Update(e.IP, e.Location)
		return nil
	}
	if mynet.CheckPrivateIPv4(e.IP) {
		// 保留地址不适合代理
		e.Location = "1;ZZ;ZZZ;Reserved"
		iPGeoCacheClient.Update(e.IP, e.Location)
		return nil
	}
	err = l.checkLocationByRemote(et, e)
	if err != nil {
		// 解析失败，尝试直连
		e.Location = "0;;;WRONG INPUT"
		iPGeoCacheClient.Delete(e.IP)
		return errors.New("Location.Send -> " + err.Error())
	}
	iPGeoCacheClient.Update(e.IP, e.Location)
	return nil
}

// Type ET子协议的类型
func (l Location) Type() int {
	return EtLOCATION
}

func (l *Location) checkLocationByRemote(et *ET, e *NetArg) error {
	req := FormatEtType(EtLOCATION) + " " + e.IP
	reply := sendQueryReq(et, req)
	e.Location = reply
	return nil
}

// Handle 处理ET-LOCATION请求
// 此方法完成缓存的读取
// 如果缓存不命中则进一步调用CheckLocationByWeb
func (l Location) Handle(req string, tunnel *mytunnel.Tunnel) error {
	reqs := strings.Split(req, " ")
	if len(reqs) < 2 {
		return errors.New("Location.Handle -> req is too short")
	}
	ip := reqs[1]

	// check by cache
	node, loaded := iPGeoCacheServer.Get(ip)
	if loaded {
		location, err := node.Wait()
		if err != nil {
			tunnel.WriteLeft([]byte(err.Error()))
			return errors.New("Location.Handle -> " + err.Error())
		}
		_, err = tunnel.WriteLeft([]byte(location))
		if err != nil {
			return errors.New("Location.Handle -> " + err.Error())
		}
		return nil
	}

	// check by web
	location, err := CheckLocationByWeb(ip)
	if err != nil {
		iPGeoCacheServer.Delete(ip)
		tunnel.WriteLeft([]byte(err.Error()))
		return errors.New("Location.Handle -> " + err.Error())
	}
	iPGeoCacheServer.Update(ip, location)
	_, err = tunnel.WriteLeft([]byte(location))
	if err != nil {
		return errors.New("Location.Handle -> " + err.Error())
	}
	return nil
}

// Match 判断业务是否匹配
func (l Location) Match(req string) bool {
	args := strings.Split(req, " ")
	if args[0] == "LOCATION" {
		return true
	}
	return false
}

// CheckProxyByLocation 本地解析IP是否需要使用代理
func (l *Location) CheckProxyByLocation(location string) bool {
	switch location {
	case "0;;;WRONG INPUT":
		return true
	case "1;ZZ;ZZZ;Reserved", l.arg.LocalLocation:
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
		return "", errors.New("CheckLocationByWeb -> " + err.Error())
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	bodyStr := string(body)
	return bodyStr, nil
}
