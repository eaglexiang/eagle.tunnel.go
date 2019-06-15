/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-13 19:04:31
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-06-15 11:53:54
 */

package cmd

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/eaglexiang/eagle.tunnel.go/src/core/protocols/et/comm"
	logger "github.com/eaglexiang/go-logger"
	mynet "github.com/eaglexiang/go-net"
	cache "github.com/eaglexiang/go-textcache"
	"github.com/eaglexiang/go-tunnel"
)

// Location ET-LOCATION子协议的实现
type Location struct {
	sync.Mutex
	cacheClient *cache.TextCache
	cacheServer *cache.TextCache
}

func (l *Location) getCacheClient(ip string) (node *cache.CacheNode, loaded bool) {
	if l.cacheClient == nil {
		l.Lock()
		if l.cacheClient == nil {
			logger.Info("create location cache for client")
			l.cacheClient = cache.CreateTextCache(0)
		}
		l.Unlock()
	}
	return l.cacheClient.Get(ip)
}

func (l *Location) getCacheServer(ip string) (node *cache.CacheNode, loaded bool) {
	if l.cacheServer == nil {
		l.Lock()
		if l.cacheServer == nil {
			logger.Info("create location cache for server")
			l.cacheServer = cache.CreateTextCache(0)
		}
		l.Unlock()
	}
	return l.cacheServer.Get(ip)
}

// Send 发送ET-LOCATION请求 解析IP的地理位置，结果存放于e.Reply
// 本方法完成缓存查询功能，查询不命中则进一步调用_Send
func (l *Location) Send(e *comm.NetArg) (err error) {
	logger.Info("resolv location for ip: ", e.IP)
	node, loaded := l.getCacheClient(e.IP)
	if loaded {
		e.Location, err = node.Wait()
	} else {
		l._Send(e, node)
	}
	logger.Info("location for ", e.IP, " is ", e.Location)
	return
}

func (l *Location) _Send(e *comm.NetArg, node *cache.CacheNode) (err error) {
	switch mynet.TypeOfAddr(e.IP) {
	case mynet.IPv6Addr:
		l.resolvIPv6(e, node)
	case mynet.IPv4Addr:
		err = l.resolvIPv4(e, node)
	default:
		logger.Warning("invalid ip: ", e.IP)
		err = errors.New("invalid ip")
	}
	return
}

func (l *Location) resolvIPv6(e *comm.NetArg, node *cache.CacheNode) {
	e.Location = "Ipv6"
	node.Update(e.Location)
	logger.Info("IPv6 found, default mode for IPv6 is proxy")
}

func (l *Location) resolvIPv4(e *comm.NetArg, node *cache.CacheNode) (err error) {
	if mynet.IsPrivateIPv4(e.IP) {
		e.Location = "1;ZZ;ZZZ;Reserved"
		node.Update(e.Location)
		logger.Info("private IPv4 found: ", e.IP)
	} else if err = l.checkLocationByRemote(e); err == nil {
		node.Update(e.Location)
		logger.Info("location for ", e.IP, ": ", e.Location)
	} else {
		e.Location = "0;;;WRONG INPUT"
		node.Destroy()
		logger.Warning(err)
	}
	return
}

// Type ET子协议的类型
func (l *Location) Type() comm.CMDType {
	return comm.LOCATION
}

// Name ET子协议的名字
func (l *Location) Name() string {
	return comm.LOCATIONTxt
}

func (l *Location) checkLocationByRemote(e *comm.NetArg) (err error) {
	e.Location, err = sendQuery(l, e.IP)
	return
}

// Handle 处理ET-LOCATION请求
// 此方法完成缓存的读取
// 如果缓存不命中则进一步调用CheckLocationByWeb
func (l *Location) Handle(req string, t *tunnel.Tunnel) (err error) {
	reqs := strings.Split(req, " ")
	if len(reqs) < 2 {
		return errors.New("Location.Handle -> req is too short")
	}
	ip := reqs[1]

	node, loaded := l.getCacheServer(ip)
	var Location string
	if loaded {
		Location, err = node.Wait()
	} else {
		Location, err = CheckLocationByWeb(ip)
		if err != nil {
			node.Destroy()
		} else {
			node.Update(Location)
		}
	}
	if err != nil {
		return err
	}
	_, err = t.WriteLeft([]byte(Location))
	return err
}

// checkProxyByLocation 本地解析IP是否需要使用代理
func checkProxyByLocation(Location string) bool {
	switch Location {
	case "0;;;WRONG INPUT":
		return true
	case "1;ZZ;ZZZ;Reserved", comm.DefaultArg.LocalLocation:
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
