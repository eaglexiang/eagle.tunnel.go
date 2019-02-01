/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-22 11:54:29
 * @LastEditTime: 2019-01-22 15:02:32
 */

package et

// Sender ET子协议的sender
type Sender interface {
	Send(et *ET, e *NetArg) error //发送流程
	ETType() int                  // ET子协议的类型
}
