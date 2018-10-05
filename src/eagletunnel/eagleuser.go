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

// EagleUser 提供基本和轻量的账户系统
type EagleUser struct {
	ID             string
	Password       string
	lastIP         string    // 上次登陆IP
	lastTime       time.Time // 上次检查登陆IP的时间
	loginMutex     sync.Mutex
	tunnels        *eaglelib.SyncList
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
func ParseEagleUser(userStr string, ip string) (*EagleUser, error) {
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
		lastIP:         ip,
		lastTime:       now,
		tunnels:        eaglelib.CreateSyncList(),
		lastCheckSpeed: now,
	}
	var pause bool
	user.pause = &pause
	if len(items) < 3 {
		return &user, nil
	}
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
	if items[3] != "" {
		var err error
		user.typeOfUser, err = parseUserType(items[3])
		if err != nil {
			return nil, err
		}
	}
	return &user, nil
}

func (user *EagleUser) toString() string {
	return user.ID + ":" + user.Password
}

// CheckAuth 检查请求EagleUser的密码是否正确，并通过校对登录地址与上次登录时间，以防止重复登录
func (user *EagleUser) CheckAuth(user2Check *EagleUser) error {
	switch user.typeOfUser {
	case PrivateUser:
		return user.checkPrivateUser(user2Check)
	case SharedUser:
		return user.checkSharedUser(user2Check)
	default:
		return errors.New("invalid user type")
	}
}

func (user *EagleUser) checkPrivateUser(user2Check *EagleUser) error {
	valid := user.Password == user2Check.Password
	if !valid {
		return errors.New("incorrent username or password")
	}
	if user.lastIP == "" {
		// 初次登录
		user.lastIP = user2Check.lastIP
		user.lastTime = user2Check.lastTime
		return nil
	}
	// 检查IP是否与上次一样
	valid = user.lastIP == user2Check.lastIP
	if valid {
		// IP相同
		return nil
	}
	user.loginMutex.Lock()
	duration := user2Check.lastTime.Sub(user.lastTime)
	valid = duration > 3*time.Minute
	if valid {
		// 3分钟内未登录过
		user.lastTime = user2Check.lastTime
		user.lastIP = user2Check.lastIP
		user.loginMutex.Unlock()
		return nil
	}
	user.loginMutex.Unlock()
	return errors.New("logined")
}

func (user *EagleUser) checkSharedUser(user2Check *EagleUser) error {
	valid := user.Password == user2Check.Password
	if !valid {
		return errors.New("incorrent username or password")
	}
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
	tunnel.Encrypt = encrypt
	tunnel.Decrypt = decrypt
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

func parseUserType(typeStr string) (int, error) {
	var err error
	var theType int
	switch typeStr {
	case "share", "SHARE", "shared", "SHARED":
		theType = SharedUser
	case "private", "PRIVATE":
		theType = PrivateUser
	default:
		err = errors.New("unknown user type")
	}
	return theType, err
}

func encrypt(data []byte) {
	for i, value := range data {
		data[i] = value ^ EncryptKey
	}
}

func decrypt(data []byte) {
	for i, value := range data {
		data[i] = value ^ EncryptKey
	}
}
