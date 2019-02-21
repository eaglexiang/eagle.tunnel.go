/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-22 11:43:46
 * @LastEditTime: 2019-02-14 12:18:30
 */

package et

import mytunnel "github.com/eaglexiang/go-tunnel"

// Handler ET子协议的Handler接口
type Handler interface {
	Handle(req string, tunnel *mytunnel.Tunnel) error // 处理业务
	Match(req string) bool                            // 判断业务是否匹配
	Type() int                                        // ET子协议的类型
}
