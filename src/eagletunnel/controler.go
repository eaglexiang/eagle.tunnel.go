/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-27 08:37:36
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-13 06:00:36
 */

package eagletunnel

import (
	"bufio"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ConfigPath 主配置文件的路径
var ConfigPath string

// ConfigKeyValues 主配置文件的所有键值对参数
var ConfigKeyValues = make(map[string]string)

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

// Timeout 超时时间（s）
var Timeout int

func init() {
	// 设定参数默认值
	ConfigKeyValues["timeout"] = "0"
	ConfigKeyValues["location"] = "1;CN;CHN;China"
	ConfigKeyValues["ip-type"] = "46"
	ConfigKeyValues["data-key"] = "34"
	ConfigKeyValues["head"] = "eagle_tunnel"
	ConfigKeyValues["proxy-status"] = "enable"
	ConfigKeyValues["user"] = "null:null"
	ConfigKeyValues["user-check"] = "off"
	ConfigKeyValues["listen"] = "0.0.0.0"
	ConfigKeyValues["relayer"] = "127.0.0.1"
	ConfigKeyValues["http"] = "off"
	ConfigKeyValues["socks"] = "off"
	ConfigKeyValues["et"] = "off"
	ConfigKeyValues["debug"] = "off"
	ConfigKeyValues["cipher"] = "simple"
}

// readConfig 读取根据给定的配置文件
func readConfig(filePath string) error {
	ConfigPath = filePath
	addDefaultArg("config-dir", filepath.Dir(ConfigPath))
	// 读取配置文件
	allConfLines, err := readLines(ConfigPath)
	if err != nil {
		return err
	}
	err = getKeyValues(ConfigKeyValues, allConfLines)
	if err != nil {
		return err
	}
	return nil
}

func addDefaultArg(key, value string) {
	if _, ok := ConfigKeyValues[key]; !ok {
		ConfigKeyValues[key] = value
	}
}

// ExecConfig 执行配置
func ExecConfig() error {
	// 读取配置文件
	if value, ok := ConfigKeyValues["config"]; ok {
		readConfig(value)
	}
	// 读取用户列表
	EnableUserCheck = ConfigKeyValues["user-check"] == "on"

	if EnableUserCheck {
		usersPath := ConfigKeyValues["config-dir"] + "/users.list"
		err := importUsers(usersPath)
		if err != nil {
			return err
		}
	}

	err := SetUser(ConfigKeyValues["user"])
	if err != nil {
		return err
	}

	EnableSOCKS5 = ConfigKeyValues["socks"] == "on"
	EnableHTTP = ConfigKeyValues["http"] == "on"
	EnableET = ConfigKeyValues["et"] == "on"

	SetListen(ConfigKeyValues["listen"])
	SetRelayer(ConfigKeyValues["relayer"])

	err = SetProxyStatus(ConfigKeyValues["proxy-status"])
	if err != nil {
		return err
	}

	if _, ok := ConfigKeyValues["config-dir"]; !ok {
		return nil
	}

	// DNS解析白名单
	whiteDomainsPath := ConfigKeyValues["config-dir"] + "/whitelist_domain.txt"
	WhitelistDomains, err = readLines(whiteDomainsPath)
	if err != nil {
		return err
	}

	// hosts文件
	err = readHosts(ConfigKeyValues["config-dir"] + "/hosts")
	if err != nil {
		return err
	}

	timeout, err := strconv.ParseInt(ConfigKeyValues["timeout"], 10, 32)
	if err != nil {
		return err
	}
	Timeout = int(timeout)

	return nil
}

//SetUser 设置本地用户
func SetUser(user string) error {
	localUser, err := ParseEagleUser(user)
	if err != nil {
		return err
	}
	LocalUser = localUser
	return err
}

//SetProxyStatus 设置Proxy-Status，enable/smart
func SetProxyStatus(status string) error {
	switch status {
	case "enable":
		ProxyStatus = ProxyENABLE
		return nil
	case "smart":
		ProxyStatus = ProxySMART
		return nil
	default:
		return errors.New("invalid value for proxy-status: " + status)
	}
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
		user, err = ParseEagleUser(line)
		if err != nil {
			return err
		}
		Users[user.ID] = user
	}
	return err
}

func getKeyValues(keyValues map[string]string, lines []string) error {
	for _, line := range lines {
		keyValue := strings.Split(line, "=")
		if len(keyValue) < 2 {
			return errors.New("invalid line: " + line)
		}
		value := keyValue[1]
		for _, item := range keyValue[2:] {
			value += "=" + item
		}
		key := strings.TrimSpace(keyValue[0])
		value = strings.TrimSpace(value)
		keyValues[key] = value
	}
	return nil
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
	items := strings.Split(localIpe, ":")
	LocalAddr = items[0]
	if len(items) >= 2 {
		LocalPort = items[1]
	} else {
		LocalPort = "8080"
	}
}

func readHosts(hostsDir string) error {

	hostsFiles := getHostsList(hostsDir)

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
			hostsCache[domain] = ip
		} else {
			return errors.New("invalid hosts line: " + host)
		}
	}
	return nil
}

func getHostsList(hostsDir string) []string {
	files, err := ioutil.ReadDir(hostsDir)
	if err != nil {
		return nil
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
	return hosts
}

//SprintConfig 将配置输出为字符串
func SprintConfig() string {
	var text string
	for k, v := range ConfigKeyValues {
		text = text + k + ": " + v + "\n"
	}
	return text
}
