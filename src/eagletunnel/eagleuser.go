/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-10-08 10:51:05
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-06 16:21:19
 */

package eagletunnel

import (
	"errors"
	"strconv"
	"strings"
)

// EagleUser 提供基本和轻量的账户系统
// 必须使用 ParseEagleUser 函数进行构造
type EagleUser struct {
	ID         string
	Password   string
	loginlog   *LoginStatus
	speedLimit uint64 // KB/s
}

// 账户类型，PrivateUser的同时登录有限制，而SharedUser则没有
const (
	SharedUser  = iota // 0 表示不限制登录
	PrivateUser        // 1 表示只允许一个登录地
)

// ParseEagleUser 通过格式化的字符串构造新的EagleUser
func ParseEagleUser(userStr string) (*EagleUser, error) {
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
	user := EagleUser{
		ID:       items[0],
		Password: items[1],
	}
	if len(items) < 3 {
		return &user, nil
	}
	// 设置限速
	if items[2] != "" {
		var err error
		user.speedLimit, err = strconv.ParseUint(items[2], 10, 64)
		if err != nil {
			return nil, errors.New("when parse EagleUser: " + err.Error())
		}
	}
	if len(items) < 4 {
		return &user, nil
	}
	// 设置最大同时登录地
	maxLoginCount := 0
	if items[3] != "" {
		var err error
		maxLoginCount, err = parseLoginCount(items[3])
		if err != nil {
			return nil, err
		}
	}
	user.loginlog = CreateLoginStatus(maxLoginCount)
	return &user, nil
}

func (user *EagleUser) toString() string {
	return user.ID + ":" + user.Password
}

// CheckAuth 检查请求EagleUser的密码是否正确，并检查是否超出登录限制
func (user *EagleUser) CheckAuth(user2Check *ReqUser) error {
	valid := user.Password == user2Check.Password
	if !valid {
		return errors.New("EagleUser.CheckAuth -> incorrent username or password")
	}
	if user.loginlog != nil {
		err := user.loginlog.Login(user2Check.IP)
		if err != nil {
			return errors.New("EagleUser.CheckAuth -> " + err.Error())
		}
	}
	return nil
}
