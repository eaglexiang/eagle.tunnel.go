package eagletunnel

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"../eaglelib/src"
)

var defaultPathsOfClientConfig = []string{
	"./client.conf",
	"./config/client.conf",
	"./eagle-tunnel.conf",
	"/etc/eagle-tunnel.d/client.conf",
}
var defaultPathsOfServerConfig = []string{
	"./server.conf",
	"./config/server.conf",
	"./eagle-tunnel.conf",
	"/etc/eagle-tunnel.d/server.conf",
}

// ConfigPath 主配置文件的路径
var ConfigPath string

// ConfigKeyValues 主配置文件的所有键值对参数
var ConfigKeyValues map[string]string

// EnableUserCheck 启用用户检查特性
var EnableUserCheck bool

// EnableSOCKS5 启用relayer对SOCKS5协议的接收
var EnableSOCKS5 bool

// EnableHTTP 启用relayer对HTTP代理协议的接收
var EnableHTTP bool

// EnableET 启用relayer对ET协议的接收
var EnableET bool

// ProxyStatus 代理的状态（全局/智能）
var ProxyStatus int

// Init 根据给定的配置文件初始化参数
func Init(filePath string) error {
	ConfigPath = filePath
	allConfLines, err := readLines(ConfigPath)
	if err != nil {
		fmt.Println("failed to read " + ConfigPath)
	}

	ConfigKeyValues = getKeyValues(allConfLines)

	// 设定默认值
	if _, ok := ConfigKeyValues["data-key"]; !ok {
		ConfigKeyValues["data-key"] = "34"
	}
	if _, ok := ConfigKeyValues["head"]; !ok {
		ConfigKeyValues["head"] = "eagle_tunnel"
	}

	if _, ok := ConfigKeyValues["config-dir"]; !ok {
		ConfigKeyValues["config-dir"] = filepath.Dir(ConfigPath)
	}

	if enableUsercheck, ok := ConfigKeyValues["user-check"]; ok {
		EnableUserCheck = enableUsercheck == "on"
	}

	if EnableUserCheck {
		usersPath := ConfigKeyValues["config-dir"] + "/users.list"
		err = importUsers(usersPath)
		if err != nil {
			fmt.Println(err)
		}
	}

	if user, ok := ConfigKeyValues["user"]; ok {
		LocalUser, err = ParseEagleUser(user, "")
		if err != nil {
			fmt.Println(err)
		}
	} else {
		LocalUser, _ = ParseEagleUser("root:root", "")
	}

	go CheckSpeedOfUsers()

	var localIpe string
	_localIpe, ok := ConfigKeyValues["listen"]
	if ok {
		localIpe = _localIpe
	}
	SetListen(localIpe)

	var socks string
	socks, ok = ConfigKeyValues["socks"]
	if ok {
		EnableSOCKS5 = socks == "on"
	}

	var http string
	http, ok = ConfigKeyValues["http"]
	if ok {
		EnableHTTP = http == "on"
	}

	var et string
	et, ok = ConfigKeyValues["et"]
	if ok {
		EnableET = et == "on"
	}

	if EnableSOCKS5 || EnableHTTP {
		var remoteIpe string
		remoteIpe, ok = ConfigKeyValues["relayer"]
		if ok {
			SetRelayer(remoteIpe)
		}
	}

	ProxyStatus = ProxyENABLE
	var status string
	status, ok = ConfigKeyValues["proxy-status"]
	if ok {
		switch status {
		case "enable":
			ProxyStatus = ProxyENABLE
		case "smart":
			ProxyStatus = ProxySMART
		default:
			ProxyStatus = ProxyENABLE
		}
	}

	whiteDomainsPath := ConfigKeyValues["config-dir"] + "/whitelist_domain.txt"
	WhitelistDomains, _ = readLines(whiteDomainsPath)

	readHosts(ConfigKeyValues["config-dir"] + "/hosts")

	return err
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
			line = strings.ToLower(line)
			lines = append(lines, line)
		}
	}
	return lines, scanner.Err()
}

func importUsers(usersPath string) error {
	Users = make(map[string]*EagleUser)
	userLines, err := readLines(usersPath)
	if err != nil {
		return err
	}
	var user *EagleUser
	for _, line := range userLines {
		user, err = ParseEagleUser(line, "")
		if err != nil {
			return err
		}
		Users[user.ID] = user
	}
	return err
}

func getKeyValues(lines []string) map[string]string {
	keyValues := make(map[string]string)
	for _, line := range lines {
		keyValue := strings.Split(line, "=")
		if len(keyValue) >= 2 {
			value := keyValue[1]
			for _, item := range keyValue[2:] {
				value += "=" + item
			}
			key := strings.TrimSpace(keyValue[0])
			value = strings.TrimSpace(value)
			keyValues[key] = value
		}
	}
	return keyValues
}

func exportKeyValues(keyValues *map[string]string, keys []string) string {
	var result string
	for _, key := range keys {
		result += key + " = " + (*keyValues)[key] + "\r\n"
	}
	return result
}

// SetRelayer 设置relayer地址
func SetRelayer(remoteIpe string) {
	items := strings.Split(remoteIpe, ":")
	RemoteAddr = strings.TrimSpace(items[0])
	if len(items) >= 2 {
		RemotePort = strings.TrimSpace(items[1])
	} else {
		RemotePort = "8080"
	}
}

// SetListen 设定本地监听地址
func SetListen(localIpe string) {
	if localIpe == "" {
		localIpe = "0.0.0.0:8080"
	}
	items := strings.Split(localIpe, ":")
	LocalAddr = items[0]
	if len(items) >= 2 {
		LocalPort = items[1]
	} else {
		LocalPort = "8080"
	}
}

func readHosts(hostsDir string) {

	hostsFiles := getHostsList(hostsDir)

	var hosts []string
	for _, file := range hostsFiles {
		newHosts, err := readLines(hostsDir + "/" + file)
		if err == nil {
			hosts = append(hosts, newHosts...)
		}
	}

	for _, host := range hosts {
		// 将所有连续两个以上空格缩减为一个
		for {
			newHost := strings.Replace(host, "  ", " ", -1)
			if newHost != host {
				host = newHost
			} else {
				break
			}
		}

		items := strings.Split(host, " ")
		if len(items) >= 2 {
			ip := strings.TrimSpace(items[0])
			domain := strings.TrimSpace(items[1])
			if domain != "" && ip != "" {
				hostsCache[domain] = ip
			}
		}
	}
}

func getHostsList(hostsDir string) []string {
	files, err := ioutil.ReadDir(hostsDir)
	if err != nil {
		return nil
	}
	var hosts []string
	for _, file := range files {
		if !file.IsDir() {
			filename := file.Name()
			if strings.HasSuffix(filename, ".hosts") {
				hosts = append(hosts, filename)
			}
		}
	}
	return hosts
}

// DefaultClientConfig 返回最匹配的client.conf文件
func DefaultClientConfig() string {
	for _, path := range defaultPathsOfClientConfig {
		if eaglelib.FileExsits(path) {
			return path
		}
	}
	return ""
}

// DefaultServerConfig 返回最匹配的server.conf文件
func DefaultServerConfig() string {
	for _, path := range defaultPathsOfServerConfig {
		if eaglelib.FileExsits(path) {
			return path
		}
	}
	return ""
}
