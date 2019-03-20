/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-03-19 20:08:49
 * @LastEditTime: 2019-03-19 20:08:51
 */

package et

import (
	"errors"

	"github.com/eaglexiang/eagle.tunnel.go/src/logger"
	mynet "github.com/eaglexiang/go-net"
	mytunnel "github.com/eaglexiang/go-tunnel"
)

// Send 发送ET请求
func (et *ET) Send(e *mynet.Arg) error {
	// 选择Sender
	newE, err := parseNetArg(e)
	if err != nil {
		return err
	}
	sender, ok := et.subSenders[newE.TheType]
	if !ok {
		logger.Error("no tcp sender")
		return errors.New("no tcp sender")
	}
	// 进入子协议业务
	err = sender.Send(et, newE)
	if err != nil {
		return err
	}
	return nil
}

func (et *ET) checkVersionOfRelayer(tunnel *mytunnel.Tunnel) error {
	req := et.arg.Head
	_, err := tunnel.WriteRight([]byte(req))
	if err != nil {
		return err
	}
	reply, err := tunnel.ReadRightStr()
	if reply != "valid valid valid" {
		logger.Warning("invalid reply for et version check: ",
			reply)
		return errors.New("invalid reply")
	}
	return nil
}

func (et *ET) checkUserOfLocal(tunnel *mytunnel.Tunnel) (err error) {
	if et.arg.LocalUser.ID == "null" {
		return nil // no need to check
	}
	user := et.arg.LocalUser.ToString()
	_, err = tunnel.WriteRight([]byte(user))
	if err != nil {
		return err
	}
	reply, err := tunnel.ReadRightStr()
	if err != nil {
		return err
	}
	if reply != "valid" {
		logger.Error("invalid reply for check local user: ", reply)
		return errors.New("invalid reply")
	}
	tunnel.SpeedLimiter = et.arg.LocalUser.SpeedLimiter()
	return nil
}

// 查询类请求的发射过程都是类似的
// 连接 - 发送请求 - 得到反馈 - 关闭连接
// 区别仅仅在请求命令的内容
func sendQueryReq(et *ET, req string) (string, error) {
	tunnel := mytunnel.GetTunnel()
	defer mytunnel.PutTunnel(tunnel)
	err := et.connect2Relayer(tunnel)
	if err != nil {
		return "", err
	}

	// 发送请求
	_, err = tunnel.WriteRight([]byte(req))
	if err != nil {
		return "", err
	}

	// 接受回复
	return tunnel.ReadRightStr()
}
