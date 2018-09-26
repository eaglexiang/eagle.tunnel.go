package eagletunnel

import (
	"container/list"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// EagleUser 提供基本和轻量的账户系统
type EagleUser struct {
	ID             string
	Password       string
	lastAddr       net.Addr  // 上次登陆地
	lastTime       time.Time // 上次检查登陆地的时间
	loginMutex     sync.Mutex
	tunnels        *SyncList
	pause          *bool
	bytes          int64
	speed          int64 // KB/s
	speedLimit     int64 // KB/s
	lastCheckSpeed time.Time
	typeOfUser     int
}

// 账户类型，PrivateUser的同时登录有限制，而SharedUser则没有
const (
	PrivateUser = iota
	SharedUser
)

// ParseEagleUser 通过格式化的字符串构造新的EagleUser，需要输入请求方地址，以防止重复登录
func ParseEagleUser(userStr string, addr net.Addr) (*EagleUser, error) {
	var user EagleUser
	var err error
	items := strings.Split(userStr, ":")
	if len(items) >= 2 {
		user = EagleUser{
			ID:             items[0],
			Password:       items[1],
			lastAddr:       addr,
			lastTime:       time.Now(),
			tunnels:        CreateSyncList(),
			lastCheckSpeed: time.Now()}
		var pause bool
		user.pause = &pause
		if len(items) >= 3 {
			user.speedLimit, err = strconv.ParseInt(items[2], 10, 64)
		}
		if len(items) >= 4 {
			switch items[3] {
			case "share":
				user.typeOfUser = SharedUser
			default:
				user.typeOfUser = PrivateUser
			}
		}
	} else {
		err = errors.New("invalid user")
	}
	return &user, err
}

func (user *EagleUser) toString() string {
	return user.ID + ":" + user.Password
}

// CheckAuth 检查请求EagleUser的密码是否正确，并通过校对登录地址与上次登录时间，以防止重复登录
func (user *EagleUser) CheckAuth(user2Check *EagleUser) error {
	switch user.typeOfUser {
	case PrivateUser:
		valid := user.Password == user2Check.Password
		if !valid {
			return errors.New("incorrent username or password")
		}
		if user.lastAddr == nil {
			user.lastAddr = user2Check.lastAddr
			user.lastTime = user2Check.lastTime
		} else {
			ip := strings.Split(user.lastAddr.String(), ":")[0]
			ip2Check := strings.Split(user2Check.lastAddr.String(), ":")[0]
			valid = ip == ip2Check
			if !valid {
				user.loginMutex.Lock()
				duration := user2Check.lastTime.Sub(user.lastTime)
				valid = duration > 3*time.Minute
				if valid {
					user.lastTime = user2Check.lastTime
					user.lastAddr = user2Check.lastAddr
					user.loginMutex.Unlock()
				} else {
					user.loginMutex.Unlock()
					return errors.New("logined")
				}
			}
		}
	case SharedUser:
		valid := user.Password == user2Check.Password
		if !valid {
			return errors.New("incorrent username or password")
		}
	}
	return nil
}

func (user *EagleUser) limitSpeed() {
	// 0表示不限速
	if user.speedLimit <= 0 {
		return
	}

	speedKB := user.speed / 1024
	*user.pause = speedKB > user.speedLimit
}

func (user *EagleUser) totalBytes() int64 {
	var totalBytes int64

	// 统计所有Tunnel的总Bytes
	var next *list.Element
	for e := user.tunnels.raw.Front(); e != nil; e = next {
		next = e.Next()
		tunnel, ok := e.Value.(*Tunnel)
		if ok {
			bytesNew := tunnel.bytesFlowed()
			totalBytes += bytesNew
			if tunnel.closed {
				user.tunnels.remove(e)
			} else {
				if tunnel.flowed && !tunnel.isRunning() {
					user.tunnels.remove(e)
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

func (user *EagleUser) addTunnel(tunnel *Tunnel) {
	tunnel.pause = user.pause
	user.tunnels.push(tunnel)
}

func (user *EagleUser) checkSpeed() {
	user.bytes += user.totalBytes()
	now := time.Now()
	duration := now.Sub(user.lastCheckSpeed)
	seconds := int64(duration.Seconds())
	if seconds > 0 {
		user.speed = user.bytes / int64(seconds)
	}
	if duration > (time.Minute * 3) {
		user.lastCheckSpeed = now
		user.bytes = 0
	}
}
