/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-10-08 10:51:05
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-03 01:37:40
 */

package eagletunnel

import (
	"container/list"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"../eaglelib/src"
)

// LoginStatus 用户的登录记录
type LoginStatus struct {
	ip       string
	lastTime time.Time
	ttl      int // min
}

func (ls *LoginStatus) isDead() bool {
	if ls.ttl == 0 {
		return false
	}
	now := time.Now()
	duration := now.Sub(ls.lastTime)
	if duration > time.Duration(ls.ttl)*time.Minute {
		return true
	}
	ls.lastTime = now
	return false
}

// EagleUser 提供基本和轻量的账户系统
type EagleUser struct {
	ID             string
	Password       string
	logins         []LoginStatus // 登录记录
	loginMutex     sync.Mutex
	tunnels        *eaglelib.SyncList
	pause          *bool
	bytes          int64
	speed          int64 // KB/s
	speedLimit     int64 // KB/s
	lastCheckSpeed time.Time
	maxLoginCount  int
}

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
		user.speedLimit, err = strconv.ParseInt(items[2], 10, 64)
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

func (user *EagleUser) limitSpeed() {
	// 0表示不限速
	if user.speedLimit == 0 {
		return
	}

	*user.pause = user.speed > user.speedLimit
}

func (user *EagleUser) totalBytes() int64 {
	var totalBytes int64

	// 统计所有Tunnel的总Bytes
	var next *list.Element
	for e := user.tunnels.Front(); e != nil; e = next {
		next = e.Next()
		tunnel, ok := e.Value.(*eaglelib.Tunnel)
		if ok {
			bytesNew := tunnel.BytesFlowed()
			totalBytes += bytesNew
			if tunnel.Closed {
				user.tunnels.Remove(e)
			} else {
				if tunnel.Flowed && !tunnel.IsRunning() {
					user.tunnels.Remove(e)
				}
			}
		} else {
			fmt.Println("error: invalid object found in EagleUser.tunnels!")
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

func (user *EagleUser) checkSpeed() {
	now := time.Now()
	duration := now.Sub(user.lastCheckSpeed)
	// 每三分钟重置一次计数
	if duration > (time.Minute * 3) {
		user.lastCheckSpeed = now
		user.bytes = 0
	}
	user.bytes += user.totalBytes()
	seconds := int64(duration.Seconds())
	if seconds > 0 {
		user.speed = user.bytes / seconds / 1024 // EagleTunnel.speed 单位为KB/s
	}
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
