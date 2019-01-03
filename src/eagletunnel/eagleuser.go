/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-10-08 10:51:05
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-03 18:07:07
 */

package eagletunnel

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"../eaglelib/src"
)

// EagleUser 提供基本和轻量的账户系统
type EagleUser struct {
	ID             string
	Password       string
	logins         []LoginStatus // 登录记录
	loginMutex     sync.Mutex
	tunnels        *eaglelib.SyncList
	pause          *bool
	speed          uint64 // Byte/s
	speedLimit     uint64 // KB/s
	lastCheckSpeed time.Time
	maxLoginCount  int
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
	now := time.Now()
	user := EagleUser{
		ID:             items[0],
		Password:       items[1],
		tunnels:        eaglelib.CreateSyncList(),
		lastCheckSpeed: now,
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
		if user.speedLimit < 0 {
			return nil, errors.New("speed limit must be bigger than or equal to 0")
		}
	}
	if len(items) < 4 {
		return &user, nil
	}
	// 设置最大同时登录地
	if items[3] != "" {
		var err error
		user.maxLoginCount, err = parseLoginCount(items[3])
		if err != nil {
			return nil, err
		}
	}
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
	if user.maxLoginCount == SharedUser {
		// 共享用户不限制登录
		return nil
	}
	// 查找该IP的既有登录
	for _, v := range user.logins {
		if v.ip == user2Check.IP {
			return nil
		}
	}
	// 查找已失效的登录
	for _, v := range user.logins {
		if v.isDead() {
			user.loginMutex.Lock()
			defer user.loginMutex.Unlock()
			v.ip = user2Check.IP
			v.lastTime = time.Now()
			return nil
		}
	}
	// 注册新登录
	if len(user.logins) >= user.maxLoginCount {
		return errors.New("too much login reqs")
	}
	user.loginMutex.Lock()
	defer user.loginMutex.Unlock()
	// 双检查，以平衡性能和线程安全
	if len(user.logins) >= user.maxLoginCount {
		return errors.New("too much login reqs")
	}
	user.logins = append(user.logins,
		LoginStatus{
			ip:       user2Check.IP,
			lastTime: time.Now(),
			ttl:      3,
		},
	)
	return nil
}

// LimitSpeed 限速
func (user *EagleUser) LimitSpeed() {
	// 0表示不限速
	if user.speedLimit == 0 {
		return
	}

	*user.pause = user.KBSpeed() > user.speedLimit
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
	now := time.Now()
	duration := now.Sub(user.lastCheckSpeed)
	user.lastCheckSpeed = now
	bytes := user.totalBytes()
	seconds := duration.Seconds()
	if seconds > 0 {
		user.speed = bytes / uint64(seconds)
	}
	user.speed = 0
}

// ByteSpeed 获取User当前的速率（Byte/s）
func (user *EagleUser) ByteSpeed() uint64 {
	return user.speed
}

// KBSpeed 获取User当前的速率（KB/s）
func (user *EagleUser) KBSpeed() uint64 {
	return user.speed / 1024
}

func parseLoginCount(arg string) (int, error) {
	switch arg {
	case "private", "PRIVATE":
		return PrivateUser, nil
	case "share", "shared", "SHARED":
		return SharedUser, nil
	default:
		value, err := strconv.ParseInt(arg, 10, 32)
		return int(value), err
	}
}
