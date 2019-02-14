/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-22 11:54:29
 * @LastEditTime: 2019-02-14 12:17:22
 */

package et

// Sender ET子协议的sender
type Sender interface {
	Send(et *ET, e *NetArg) error //发送流程
	Type() int                    // ET子协议的类型
}
