/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-13 19:04:31
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-03-17 17:32:04
 */

package et

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	logger "github.com/eaglexiang/eagle.tunnel.go/src/logger"
	mynet "github.com/eaglexiang/go-net"
	cache "github.com/eaglexiang/go-textcache"
	mytunnel "github.com/eaglexiang/go-tunnel"
)

// Location ET-LOCATION子协议的实现
type Location struct {
	arg         *Arg
	cacheClient *cache.TextCache
	cacheServer *cache.TextCache
}

func (l *Location) getCacheClient(ip string) (node *cache.CacheNode, loaded bool) {
	if l.cacheClient == nil {
		l.cacheClient = cache.CreateTextCache(0)
	}
	return l.cacheClient.Get(ip)
}

func (l *Location) getCacheServer(ip string) (node *cache.CacheNode, loaded bool) {
	if l.cacheServer == nil {
		l.cacheServer = cache.CreateTextCache(0)
	}
	return l.cacheServer.Get(ip)
}

// Send 发送ET-LOCATION请求 解析IP的地理位置，结果存放于e.Reply
// 本方法完成缓存查询功能，查询不命中则进一步调用_Send
func (l Location) Send(et *ET, e *NetArg) (err error) {
	node, loaded := l.getCacheClient(e.IP)
	if loaded {
		e.Location, err = node.Wait()
	} else {
		l._Send(et, e, node)
	}
	return
}

func (l Location) _Send(et *ET, e *NetArg, node *cache.CacheNode) (err error) {
	switch mynet.TypeOfAddr(e.IP) {
	case mynet.IPv6Addr:
		// IPv6 默认代理
		e.Location = "Ipv6"
		node.Update(e.Location)
	case mynet.IPv4Addr:
		if mynet.CheckPrivateIPv4(e.IP) {
			// 保留地址不适合代理
			e.Location = "1;ZZ;ZZZ;Reserved"
			node.Update(e.Location)
		} else if err = l.checkLocationByRemote(et, e); err == nil {
			node.Update(e.Location)
		} else {
			e.Location = "0;;;WRONG INPUT"
			l.cacheClient.Delete(e.IP)
		}
	default:
		logger.Warning("invalid ip: ", e.IP)
		err = errors.New("invalid ip")
	}
	return
}

// Type ET子协议的类型
func (l Location) Type() int {
	return EtLOCATION
}

func (l *Location) checkLocationByRemote(et *ET, e *NetArg) (err error) {
	req := FormatEtType(EtLOCATION) + " " + e.IP
	e.Location, err = sendQueryReq(et, req)
	return
}

// Handle 处理ET-LOCATION请求
// 此方法完成缓存的读取
// 如果缓存不命中则进一步调用CheckLocationByWeb
func (l Location) Handle(req string, tunnel *mytunnel.Tunnel) (err error) {
	reqs := strings.Split(req, " ")
	if len(reqs) < 2 {
		return errors.New("Location.Handle -> req is too short")
	}
	ip := reqs[1]

	node, loaded := l.getCacheServer(ip)
	var location string
	if loaded {
		location, err = node.Wait()
	} else {
		location, err = CheckLocationByWeb(ip)
		if err != nil {
			l.cacheServer.Delete(ip)
		} else {
			node.Update(location)
		}
	}
	if err != nil {
		return err
	}
	_, err = tunnel.WriteLeft([]byte(location))
	return err
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
		logger.Warning(err)
		return "", err
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	bodyStr := string(body)
	return bodyStr, nil
}
