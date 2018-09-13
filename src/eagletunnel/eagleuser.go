package eagletunnel

import (
	"errors"
	"net"
	"strings"
	"sync"
	"time"
)

type EagleUser struct {
	Id         string
	Password   string
	lastAddr   net.Addr
	lastTime   time.Time
	loginMutex sync.Mutex
}

func ParseEagleUser(userStr string, addr net.Addr) (*EagleUser, error) {
	var user EagleUser
	var err error
	items := strings.Split(userStr, ":")
	if len(items) >= 2 {
		user = EagleUser{Id: items[0], Password: items[1], lastAddr: addr, lastTime: time.Now()}
	} else {
		err = errors.New("invalid user")
	}
	return &user, err
}

func (user *EagleUser) toString() string {
	return user.Id + ":" + user.Password
}

func (user *EagleUser) Check(user2Check *EagleUser) error {
	valid := user.Password == user2Check.Password
	if !valid {
		return errors.New("incorrent password")
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
	return nil
}
