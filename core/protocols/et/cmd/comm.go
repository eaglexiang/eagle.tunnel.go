/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-08-24 10:48:07
 * @LastEditTime: 2019-08-24 10:48:51
 */

package cmd

import "github.com/eaglexiang/eagle.tunnel.go/core/protocols/et/comm"

func sendQuery(s comm.Sender, req string) (string, error) {
	req = s.Name() + " " + req
	return comm.SendQueryReq(req)
}
