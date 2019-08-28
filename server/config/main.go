/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-08-24 10:52:26
 * @LastEditTime: 2019-08-28 20:50:35
 */

package config

import (
	"github.com/eaglexiang/go/logger"
	"github.com/eaglexiang/go/settings"
	myuser "github.com/eaglexiang/go/user"
)

// ImportConfigFiles 导入配置
func ImportConfigFiles() {
	readConfigFile()

	initLogger()
	initListens()
	initRelays()
	initProxyStatus()
	initLocalUser()
	initTimeout()
	initBufferSize()

	readConfigDir()
}

// readConfigFile 读取根据给定的配置文件
func readConfigFile() {
	if !settings.Exsit("config") {
		return
	}

	filePath := settings.Get("config")
	allConfLines := readLinesFromFile(filePath)
	err := settings.ImportLines(allConfLines)
	if err != nil {
		panic(err)
	}
}

//SetUser 设置本地用户
func SetUser(user string) {
	var err error
	LocalUser, err = myuser.ParseValidUser(user)
	if err != nil {
		panic(err)
	}
}

func importUsers(usersPath string) {
	Users = make(map[string]*myuser.ValidUser)
	userLines := readLinesFromFile(usersPath)

	for _, line := range userLines {
		user, err := myuser.ParseValidUser(line)
		if err != nil {
			panic(err)
		}
		Users[user.ID] = user
	}
	logger.Info(len(Users), " users imported")
}
