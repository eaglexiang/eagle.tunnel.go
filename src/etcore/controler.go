/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-27 08:37:36
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-02-17 00:58:06
 */

package etcore

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"plugin"
	"strconv"
	"strings"

	"github.com/eaglexiang/go-settings"
	myuser "github.com/eaglexiang/go-user"

	myet "./et"
)

// ConfigPath 主配置文件的路径
var ConfigPath string

// EnableSOCKS5 启用relayer对SOCKS5协议的接收
var EnableSOCKS5 bool

// EnableHTTP 启用relayer对HTTP代理协议的接收
var EnableHTTP bool

// EnableET 启用relayer对ET协议的接收
var EnableET bool

// ProxyStatus 代理的状态（全局/智能）
var ProxyStatus int

// Timeout 超时时间（s）
var Timeout int

func init() {
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
	settings.SetDefault("relayer", "127.0.0.1")
	settings.SetDefault("http", "off")
	settings.SetDefault("socks", "off")
	settings.SetDefault("et", "off")
	settings.SetDefault("debug", "off")
	settings.SetDefault("cipher", "simple")
}

// readConfig 读取根据给定的配置文件
func readConfig(filePath string) error {
	ConfigPath = filePath
	settings.SetDefault("config-dir", filepath.Dir(ConfigPath))
	// 读取配置文件
	allConfLines, err := readLines(ConfigPath)
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
func ExecConfig() error {
	// 读取配置文件
	if settings.Exsit("config") {
		readConfig(settings.Get("config"))
	}
	// 读取用户列表
	if settings.Get("user-check") == "on" {
		usersPath := settings.Get("config-dir") + "/users.list"
		err := importUsers(usersPath)
		if err != nil {
			return err
		}
	}

	// 读取本地用户
	err := SetUser(settings.Get("user"))
	if err != nil {
		return err
	}

	EnableSOCKS5 = settings.Get("socks") == "on"
	EnableHTTP = settings.Get("http") == "on"
	EnableET = settings.Get("et") == "on"

	SetListen(settings.Get("listen"))
	SetRelayer(settings.Get("relayer"))

	err = SetProxyStatus(settings.Get("proxy-status"))
	if err != nil {
		return err
	}

	if !settings.Exsit("config-dir") {
		return nil
	}

	// DNS解析白名单
	whiteDomainsPath := settings.Get("config-dir") + "/whitelist_domain.txt"
	myet.WhitelistDomains, _ = readLines(whiteDomainsPath)

	// hosts文件
	if settings.Exsit("config-dir") {
		hostsDir := settings.Get("config-dir") + "/hosts"
		err = readHosts(hostsDir)
		if err != nil {
			return errors.New("ExecConfig -> " + err.Error())
		}
	}

	timeout, err := strconv.ParseInt(
		settings.Get("timeout"),
		10,
		32)
	if err != nil {
		return errors.New("ExecConfig -> " + err.Error())
	}
	Timeout = int(timeout)

	SetDebug(settings.Get("debug"))

	// 导入Mods
	if settings.Exsit("mod-dir") {
		modsDir := settings.Get("mod-dir")
		err = ImportMods(modsDir)
		if err != nil {
			return errors.New("ExecConfig -> " + err.Error())
		}
	}

	return nil
}

//SetUser 设置本地用户
func SetUser(user string) error {
	localUser, err := myuser.ParseUser(user)
	if err != nil {
		return err
	}
	LocalUser = localUser
	return err
}

//SetProxyStatus 设置Proxy-Status，enable/smart
func SetProxyStatus(status string) error {
	ProxyStatus = myet.ParseProxyStatus(status)
	if ProxyStatus == myet.ErrorProxyStatus {
		return errors.New("SetProxyStatus -> invalid proxy-status")
	}
	settings.Set("proxy-status", status)
	return nil
}

// SetDebug 设置Debug
func SetDebug(debug string) {
	if debug == "on" {
		Debug = true
	}
	settings.Set("debug", debug)
}

func readLines(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
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

func importUsers(usersPath string) error {
	Users = make(map[string]*myuser.User)
	userLines, err := readLines(usersPath)
	if err != nil {
		return nil
	}
	var user *myuser.User
	for _, line := range userLines {
		user, err = myuser.ParseUser(line)
		if err != nil {
			return err
		}
		Users[user.ID()] = user
	}
	return err
}

// SetRelayer 设置relayer地址
func SetRelayer(remoteIpe string) {
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
	settings.Set("relayer", remoteIpe)
}

// SetListen 设定本地监听地址
func SetListen(localIpe string) {
	if strings.HasPrefix(localIpe, "[") {
		// IPv6
		if strings.HasSuffix(localIpe, "]") {
			// 不包含端口号
			localIpe += ":8080"
		}
	} else {
		ip := net.ParseIP(localIpe)
		if ip != nil {
			// 不包含端口号
			localIpe += ":8080"
		}
	}
	settings.Set("listen", localIpe)
}

func readHosts(hostsDir string) error {

	hostsFiles, err := getHostsList(hostsDir)
	if err != nil {
		return errors.New("readHosts -> " +
			err.Error())
	}

	var hosts []string
	for _, file := range hostsFiles {
		newHosts, err := readLines(hostsDir + "/" + file)
		if err != nil {
			return err
		}
		hosts = append(hosts, newHosts...)
	}

	for _, host := range hosts {
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
			return errors.New("invalid hosts line: " + host)
		}
		ip := strings.TrimSpace(items[0])
		domain := strings.TrimSpace(items[1])
		if domain != "" && ip != "" {
			myet.HostsCache[domain] = ip
		} else {
			return errors.New("invalid hosts line: " + host)
		}
	}
	return nil
}

func getHostsList(hostsDir string) ([]string, error) {
	files, err := ioutil.ReadDir(hostsDir)
	if err != nil {
		return nil, errors.New("getHostsList -> " +
			err.Error())
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
		return errors.New("ImportMods -> " +
			err.Error())
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
				fmt.Println(err)
			}
		}
	}
	return nil
}
