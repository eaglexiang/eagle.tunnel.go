/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-02 15:40:22
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-03-03 05:31:46
 */

package et

import (
	"github.com/eaglexiang/go-cache"
)

// LocationCache IP-GEO缓存
type LocationCache struct {
	data *eaglecache.Cache
}

// CreateLocationCache 创建LocationCache
func CreateLocationCache() *LocationCache {
	return &LocationCache{
		data: eaglecache.CreateCache(0), // 记录不过期，不进行过期检查
	}
}

// Get 获取指定IP的节点
// 以及该节点之前是否存在
func (cache *LocationCache) Get(ip string) (node *LocationCacheNode, loaded bool) {
	var _node *eaglecache.CacheNode
	_node, loaded = cache.data.Get(ip)
	node = &LocationCacheNode{node: _node}
	return
}

// Update 更新IP记录的值，并解除阻塞
func (cache *LocationCache) Update(ip string, location string) {
	cache.data.Update(ip, location)
}

// Delete 删除IP的记录
func (cache *LocationCache) Delete(ip string) {
	cache.data.Delete(ip)
}
