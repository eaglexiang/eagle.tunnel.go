/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-27 08:37:36
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-03-17 16:39:08
 */

package core

import (
	"bufio"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"plugin"
	"strconv"
	"strings"
	"time"

	"github.com/eaglexiang/eagle.tunnel.go/src/core/protocols/et/comm"
	"github.com/eaglexiang/eagle.tunnel.go/src/logger"
	settings "github.com/eaglexiang/go-settings"
	myuser "github.com/eaglexiang/go-user"
)

// ConfigPath 主配置文件的路径
var ConfigPath string

// ProxyStatus 代理的状态（全局/智能）
var ProxyStatus int

// Timeout 超时时间
var Timeout time.Duration

func init() {
	settings.Bind("relayer", "relay")
	// 设定参数默认值
	settings.SetDefault("timeout", "0")
	settings.SetDefault("location", "1;CN;CHN;China")
	settings.SetDefault("ip-type", "46")
	settings.SetDefault("data-key", "34")
	settings.SetDefault("head", "eagle_tunnel")
	settings.SetDefault("proxy-status", "enable")
	settings.SetDefault("user", "null:null")
	settings.SetDefault("user-check", "off")
	settings.SetDefault("listen", "0.0.0.0")
	settings.SetDefault("relay", "127.0.0.1")
	settings.SetDefault("http", "off")
	settings.SetDefault("socks", "off")
	settings.SetDefault("et", "off")
	settings.SetDefault("debug", "warning")
	settings.SetDefault("cipher", "simple")
	settings.SetDefault("maxclients", "0")
}

// readConfig 读取根据给定的配置文件
func readConfig() error {
	if !settings.Exsit("config") {
		return nil
	}
	filePath := settings.Get("config")
	allConfLines, err := readLines(filePath)
	if err != nil {
		return err
	}
	err = settings.ImportLines(allConfLines)
	if err != nil {
		return err
	}
	return nil
}

// ExecConfig 执行配置
func ExecConfig() (err error) {
	err = readConfig()
	if err != nil {
		return err
	}

	settings.Set("listen", SetIPE(settings.Get("listen")))
	settings.Set("relay", SetIPE(settings.Get("relay")))

	err = SetProxyStatus(settings.Get("proxy-status"))
	if err != nil {
		return err
	}

	return readConfigDir()
}

func readConfigDir() (err error) {
	if !finishConfigDir() {
		return nil
	}
	err = execUserSystem()
	if err != nil {
		return err
	}
	// DNS解析白名单
	whiteDomainsPath := settings.Get("config-dir") + "/whitelist_domain.txt"
	comm.WhitelistDomains, err = readLines(whiteDomainsPath)
	if err != nil {
		return
	}
	// hosts文件
	err = execHosts()
	if err != nil {
		return
	}
	execTimeout()
	// 导入Mods
	return execMods()
}

// finishConfigDir 补全config-dir
// 返回值表示是否有可用的config-dir
func finishConfigDir() bool {
	if !settings.Exsit("config-dir") {
		if !settings.Exsit("config") {
			return false
		}
		settings.Set("config-dir", filepath.Dir(settings.Get("config")))
	}
	return true
}

func execUserSystem() (err error) {
	// 读取用户列表
	if settings.Get("user-check") == "on" {
		usersPath := settings.Get("config-dir") + "/users.list"
		err = importUsers(usersPath)
		if err != nil {
			return
		}
	}

	// 读取本地用户
	if !settings.Exsit("user") {
		SetUser("null:null")
	} else {
		err = SetUser(settings.Get("user"))
	}
	return err
}

func execTimeout() error {
	_timeout := settings.Get("timeout")
	timeout, err := strconv.ParseInt(
		_timeout,
		10,
		32)
	if err != nil {
		logger.Error("invalid timeout", _timeout)
		return err
	}
	Timeout = time.Second * time.Duration(timeout)
	return nil
}

func execHosts() (err error) {
	hostsDir := settings.Get("config-dir") + "/hosts"
	err = readHosts(hostsDir)
	if err != nil {
		return err
	}
	return nil
}

func execMods() (err error) {
	if modsDir := settings.Get("mod-dir"); modsDir != "" {
		err = ImportMods(modsDir)
		if err != nil {
			return err
		}
	}
	return nil
}

//SetUser 设置本地用户
func SetUser(user string) (err error) {
	LocalUser, err = myuser.ParseValidUser(user)
	return
}

//SetProxyStatus 设置Proxy-Status，enable/smart
func SetProxyStatus(status string) (err error) {
	ProxyStatus, err = comm.ParseProxyStatus(status)
	if err != nil {
		logger.Error(err)
	}
	return
}

func readLines(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		items := strings.Split(line, "#")
		line = strings.TrimSpace(items[0])
		if line != "" {
			line = strings.Replace(line, "\t", " ", -1)
			lines = append(lines, line)
		}
	}
	return lines, scanner.Err()
}

func importUsers(usersPath string) (err error) {
	Users = make(map[string]*myuser.ValidUser)
	userLines, err := readLines(usersPath)
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

// SetIPE 补全端口号
func SetIPE(remoteIpe string) string {
	if strings.HasPrefix(remoteIpe, "[") {
		// IPv6
		if strings.HasSuffix(remoteIpe, "]") {
			// 不包含端口号
			remoteIpe += ":8080"
		}
	} else {
		ip := net.ParseIP(remoteIpe)
		if ip != nil {
			// 不包含端口号
			remoteIpe += ":8080"
		}
	}
	return remoteIpe
}

func readHosts(hostsDir string) error {
	hostsFiles, err := getHostsList(hostsDir)
	if err != nil {
		return err
	}

	var hosts []string
	for _, file := range hostsFiles {
		newHosts, err := readLines(hostsDir + "/" + file)
		if err != nil {
			return err
		}
		hosts = append(hosts, newHosts...)
	}

	for index, host := range hosts {
		err = handleSingleHost(host)
		if err != nil {
			return err
		}
		hosts[index] = host
	}
	return nil
}

func handleSingleHost(host string) (err error) {
	// 将所有连续空格缩减为单个空格
	for {
		newHost := strings.Replace(host, "  ", " ", -1)
		if newHost == host {
			break
		}
		host = newHost
	}

	items := strings.Split(host, " ")
	if len(items) < 2 {
		panic("invalid hosts line: " + host)
	}
	ip := strings.TrimSpace(items[0])
	domain := strings.TrimSpace(items[1])
	if domain != "" && ip != "" {
		comm.HostsCache[domain] = ip
	} else {
		panic("invalid hosts line: " + host)
	}
	return nil
}

func getHostsList(hostsDir string) ([]string, error) {
	files, err := ioutil.ReadDir(hostsDir)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	var hosts []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filename := file.Name()
		if strings.HasSuffix(filename, ".hosts") {
			hosts = append(hosts, filename)
		}
	}
	return hosts, nil
}

// ImportMods 导入Mods
func ImportMods(modsDir string) error {
	files, err := ioutil.ReadDir(modsDir)
	if err != nil {
		logger.Error(err)
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filename := file.Name()
		filename = modsDir + "/" + filename
		if strings.HasSuffix(filename, ".so") {
			_, err := plugin.Open(filename)
			if err != nil {
				logger.Error(err)
				return err
			}
		}
	}
	return nil
}
