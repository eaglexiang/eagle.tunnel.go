/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-04-03 20:30:09
 * @LastEditTime: 2019-04-03 20:30:14
 */
package comm

import (
	"github.com/eaglexiang/go-tunnel"
)

// Connect2Remote 连接到远端
var Connect2Remote func(tunnel *tunnel.Tunnel) error

// SendQueryReq 发送查询请求
// 查询类请求的发射过程都是类似的
// 连接 - 发送请求 - 得到反馈 - 关闭连接
// 区别仅仅在请求命令的内容
func SendQueryReq(req string) (reply string, err error) {
	t := tunnel.GetTunnel()
	defer tunnel.PutTunnel(t)
	t.Update(tunnel.WithTimeout(Timeout))
	err = Connect2Remote(t)
	if err != nil {
		return
	}

	// 发送请求
	_, err = t.WriteRight([]byte(req))
	if err != nil {
		return "", err
	}

	// 接受回复
	return t.ReadRightStr()
}
