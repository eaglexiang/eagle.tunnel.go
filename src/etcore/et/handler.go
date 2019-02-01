/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-22 11:43:46
 * @LastEditTime: 2019-01-22 13:05:51
 */

package et

import mytunnel "github.com/eaglexiang/go-tunnel"

// Handler ET子协议的Handler接口
type Handler interface {
	Handle(req string, tunnel *mytunnel.Tunnel) error
	Match(req string) bool
}
