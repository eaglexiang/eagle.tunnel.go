package eagletunnel

import (
	"../eaglelib/src"
)

// ProxyCache 记录是否使用代理
type ProxyCache struct {
	data *eaglelib.Cache
}

// CreateProxyCache 创建LocationCache
func CreateProxyCache() *ProxyCache {
	data := eaglelib.CreateCache(-1, 0) // 记录不过期，不进行过期检查
	cache := ProxyCache{data: data}
	return &cache
}

// Add 添加IP记录
func (cache *ProxyCache) Add(ip string) {
	cache.data.Add(ip)
}

// Update 更新IP记录的值，并解除阻塞
func (cache *ProxyCache) Update(ip string, proxy bool) {
	cache.data.Update(ip, proxy)
}

// Delete 删除IP的记录
func (cache *ProxyCache) Delete(ip string) {
	cache.data.Delete(ip)
}

// Exsit 判断IP的记录是否存在
func (cache *ProxyCache) Exsit(ip string) bool {
	return cache.data.Exsit(ip)
}

// Wait4Proxy 等待IP是否需要代理的解析结果
func (cache *ProxyCache) Wait4Proxy(ip string) (proxy bool, err error) {
	_proxy, err := cache.data.Wait(ip)
	if err != nil {
		return false, err
	}
	proxy = _proxy.(bool)
	return proxy, nil
}
