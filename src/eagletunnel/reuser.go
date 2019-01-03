/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-03 18:07:15
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-03 18:07:47
 */

package eagletunnel

import (
	"errors"
	"strings"
)

// ReqUser 请求登录使用的临时用户
type ReqUser struct {
	ID       string
	Password string
	IP       string
}

// ParseReqUser 通过字符串创建ReqUser
func ParseReqUser(userStr, ip string) (*ReqUser, error) {
	items := strings.Split(userStr, ":")
	if len(items) < 2 {
		return nil, errors.New("invalid user text")
	}
	if items[0] == "" {
		return nil, errors.New("null username")
	}
	if items[1] == "" {
		return nil, errors.New("null password")
	}
	user := ReqUser{
		ID:       items[0],
		Password: items[1],
		IP:       ip,
	}
	return &user, nil
}
