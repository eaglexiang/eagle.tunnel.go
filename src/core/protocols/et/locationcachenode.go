/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-03-03 05:27:00
 * @LastEditTime: 2019-03-03 05:29:33
 */

package et

import (
	"errors"

	mycache "github.com/eaglexiang/go-cache"
)

// LocationCacheNode LocationCache使用的节点
// 用来等待Location请求的返回
// 以实现请求结果的复用
type LocationCacheNode struct {
	node *mycache.CacheNode
}

// Wait 等待LOCATION解析请求的返回
func (node *LocationCacheNode) Wait() (location string, err error) {
	v, err := node.node.Wait4Value()
	if err != nil {
		return "", errors.New("LocationCacheNode.Wait -> " +
			err.Error())
	}
	return v.(string), nil
}
