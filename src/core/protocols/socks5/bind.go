/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-02-24 18:40:56
 * @LastEditTime: 2019-02-24 19:00:10
 */

package socks5

import mynet "github.com/eaglexiang/go-net"

type bind struct {
}

func (b bind) Handle(req []byte, e *mynet.Arg) error {
	return nil
}
