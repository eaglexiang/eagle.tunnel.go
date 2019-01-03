/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-10-08 10:51:05
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-03 23:51:43
 */

package eagletunnel

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"../eaglelib/src"
)

// EagleUser 提供基本和轻量的账户系统
// 必须使用 ParseEagleUser 函数进行构造
type EagleUser struct {
	ID                string
	Password          string
	loginlog          *LoginStatus
	tunnels           *eaglelib.SyncList
	pause             *bool
	bytes             eaglelib.AsyncCounter // Byte
	speed             uint64                // Byte/s
	speedLimit        uint64                // KB/s
	checkSpeedFrom    time.Time             // 从这个时间开始阻塞
	currentCheckSpeed time.Time             // 上次检查速度的时间
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
		tunnels:  eaglelib.CreateSyncList(),
		bytes:    eaglelib.CreateAsyncCounter(),
	}
	var pause bool
	user.pause = &pause
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
		return errors.New("incorrent username or password")
	}
	if user.loginlog != nil {
		return user.loginlog.Login(user2Check.IP)
	}
	return nil
}

func (user *EagleUser) totalBytes() uint64 {
	var totalBytes uint64

	// 统计所有Tunnel的总Bytes
	for e := user.tunnels.Front(); e != nil; e = e.Next() {
		tunnel, ok := e.Value.(*eaglelib.Tunnel)
		if !ok {
			fmt.Println("error: invalid object found in EagleUser.tunnels!")
			continue
		}
		bytesNew := tunnel.BytesFlowed()
		totalBytes += bytesNew
		if tunnel.Closed() {
			user.tunnels.Remove(e)
			continue
		}
	}

	if totalBytes < 0 {
		totalBytes = 0
	}
	return totalBytes
}

func (user *EagleUser) addTunnel(tunnel *eaglelib.Tunnel) {
	tunnel.Pause = user.pause
	user.tunnels.Push(tunnel)
}

// CheckSpeed 该User当前的速率（Byte/s）
func (user *EagleUser) CheckSpeed() {
	user.currentCheckSpeed = time.Now()
	duration := user.currentCheckSpeed.Sub(user.checkSpeedFrom)
	seconds := uint64(duration.Seconds())
	user.bytes.Add(user.totalBytes())
	if seconds > 0 {
		user.speed = user.bytes.Value() / seconds
	}
}

// ByteSpeed 获取User当前的速率（Byte/s）
func (user *EagleUser) ByteSpeed() uint64 {
	return user.speed
}

// KBSpeed 获取User当前的速率（KB/s）
func (user *EagleUser) KBSpeed() uint64 {
	return user.speed / 1024
}

// LimitSpeed 限速
func (user *EagleUser) LimitSpeed() {
	if user.speedLimit == 0 {
		// 0表示不限速
		*user.pause = false
		user.checkSpeedFrom = user.currentCheckSpeed
		user.bytes.Set(0)
		return
	}
	if user.KBSpeed() > user.speedLimit {
		*user.pause = true
	} else {
		*user.pause = false
		user.checkSpeedFrom = user.currentCheckSpeed
		user.bytes.Set(0)
	}
}
