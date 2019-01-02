/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-02 15:40:22
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-02 15:49:49
 */

package eagletunnel

import (
	"../eaglelib/src"
)

// LocationCache IP-GEO缓存
type LocationCache struct {
	data *eaglelib.Cache
}

// CreateLocationCache 创建LocationCache
func CreateLocationCache() *LocationCache {
	data := eaglelib.CreateCache(-1, 0) // 记录不过期，不进行过期检查
	cache := LocationCache{data: data}
	return &cache
}

// Add 添加IP记录
func (cache *LocationCache) Add(ip string) {
	cache.data.Add(ip)
}

// Update 更新IP记录的值，并解除阻塞
func (cache *LocationCache) Update(ip string, location string) {
	cache.data.Update(ip, location)
}

// Delete 删除IP的记录
func (cache *LocationCache) Delete(ip string) {
	cache.data.Delete(ip)
}

// Exsit 判断IP的记录是否存在
func (cache *LocationCache) Exsit(ip string) bool {
	return cache.data.Exsit(ip)
}

// Wait4Proxy 等待IP的值
func (cache *LocationCache) Wait4Proxy(ip string) (location string, err error) {
	_location, err := cache.data.Wait(ip)
	if err != nil {
		return "error", err
	}
	location = _location.(string)
	return location, nil
}
