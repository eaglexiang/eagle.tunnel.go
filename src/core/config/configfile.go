package config

import (
	"net"
	"path"
	"strings"
	"time"

	"github.com/eaglexiang/eagle.tunnel.go/src/core/protocols/et/comm"
	"github.com/eaglexiang/go-bytebuffer"
	"github.com/eaglexiang/go-logger"
	mynet "github.com/eaglexiang/go-net"
	"github.com/eaglexiang/go-settings"
	myuser "github.com/eaglexiang/go-user"
)

// ImportConfigFile 导入配置
func ImportConfigFile() (err error) {
	if err = readConfigFile(); err != nil {
		return
	}

	settings.Set("listen", finishPort(settings.Get("listen")))
	settings.Set("relay", finishPort(settings.Get("relay")))

	if err = SetProxyStatus(settings.Get("proxy-status")); err != nil {
		return
	}
	if err = initLocalUser(); err != nil {
		return
	}
	initTimeout()
	initBufferSize()
	return readConfigDir()
}

// readConfigFile 读取根据给定的配置文件
func readConfigFile() error {
	if !settings.Exsit("config") {
		return nil
	}
	filePath := settings.Get("config")
	allConfLines, err := readLinesFromFile(filePath)
	if err != nil {
		return err
	}
	return settings.ImportLines(allConfLines)
}

func initTimeout() {
	timeout := settings.GetInt64("timeout")
	Timeout = time.Second * time.Duration(timeout)
	comm.Timeout = Timeout
}

func initBufferSize() {
	size := settings.GetInt64("buffer.size")
	bytebuffer.SetDefaultSize(int(size))
}

func initLocalUser() (err error) {
	// 读取本地用户
	if !settings.Exsit("user") {
		SetUser("null:null")
	} else {
		err = SetUser(settings.Get("user"))
	}
	return
}

// initUserList 初始化用户列表
func initUserList() (err error) {
	if settings.Get("user-check") == "on" {
		usersPath := path.Join(settings.Get("config-dir"), "/users.list")
		err = importUsers(usersPath)
		if err != nil {
			return
		}
	}
	return err
}

//SetUser 设置本地用户
func SetUser(user string) (err error) {
	LocalUser, err = myuser.ParseValidUser(user)
	return
}

func importUsers(usersPath string) (err error) {
	Users = make(map[string]*myuser.ValidUser)
	userLines, err := readLinesFromFile(usersPath)
	if err != nil {
		return nil
	}
	var user *myuser.ValidUser
	for _, line := range userLines {
		user, err = myuser.ParseValidUser(line)
		if err != nil {
			return err
		}
		Users[user.ID] = user
	}
	logger.Info(len(Users), " users imported")
	return
}

// finishPort 补全端口号
func finishPort(remoteIpe string) string {
	switch mynet.TypeOfAddr(remoteIpe) {
	case mynet.IPv4Addr:
		if ip := net.ParseIP(remoteIpe); ip != nil {
			// 不包含端口号
			remoteIpe += ":8080"
		}
	case mynet.IPv6Addr:
		if strings.HasSuffix(remoteIpe, "]") {
			// 不包含端口号
			remoteIpe += ":8080"
		}
	}
	return remoteIpe
}

//SetProxyStatus 设置Proxy-Status，enable/smart
func SetProxyStatus(status string) (err error) {
	ProxyStatus, err = comm.ParseProxyStatus(status)
	if err != nil {
		logger.Error(err)
	}
	return
}
