/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-04 14:46:10
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-04 18:40:23
 */

package eagletunnel

// RemoteAddr 上级Relayer地址
var RemoteAddr string

// RemotePort 上级Relayer端口
var RemotePort string

// Sender 请求发送者
type Sender interface {
	Send(e NetArg) error
}
