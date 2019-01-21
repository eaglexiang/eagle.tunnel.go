/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-04 14:46:00
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-21 17:36:54
 */

package eagletunnel

import mytunnel "github.com/eaglexiang/go-tunnel"

// Handler 请求处理者
type Handler interface {
	Handle(request Request, tunnel *mytunnel.Tunnel) error
}
