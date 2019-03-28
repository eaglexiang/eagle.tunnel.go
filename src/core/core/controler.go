/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-27 08:37:36
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-03-29 00:14:19
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
	settings.Bind("channel", "data-key")
	// 设定参数默认值
	settings.SetDefault("timeout", "30")
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
	allConfLines, err := readLinesFromFile(filePath)
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
		return
	}

	settings.Set("listen", SetIPE(settings.Get("listen")))
	settings.Set("relay", SetIPE(settings.Get("relay")))

	err = SetProxyStatus(settings.Get("proxy-status"))
	if err != nil {
		return
	}

	err = initLocalUser()
	if err != nil {
		return
	}

	return readConfigDir()
}

func readConfigDir() (err error) {
	if !finishConfigDir() {
		return nil
	}
	err = initUserList()
	if err != nil {
		return err
	}
	err = readClearDomains()
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

// readClearDomains 读取明确连接规则的域名列表
func readClearDomains() error {
	if err := readDirectlist(); err != nil {
		return err
	}
	return readProxylist()
}

// readWhitelist 读取强制代理的域名列表
func readProxylist() (err error) {
	proxyDomainsPath := settings.Get("config-dir") + "/proxylists"
	proxyDomains, err := readLinesFromDir(proxyDomainsPath, ".txt")
	if err != nil {
		return err
	}
	for _, domain := range proxyDomains {
		comm.ProxyDomains.ReverseGrow(domain)
	}
	logger.Info(comm.ProxyDomains.Count(), " proxy-domains imported")
	return
}

// readDirectlist 读取强制直连的域名列表
func readDirectlist() (err error) {
	directDomainsPath := settings.Get("config-dir") + "/directlists"
	directDomains, err := readLinesFromDir(directDomainsPath, ".txt")
	if err != nil {
		return err
	}
	for _, domain := range directDomains {
		comm.DirectDomains.ReverseGrow(domain)
	}
	logger.Info(comm.DirectDomains.Count(), " direct-domains imported")
	return
}

// readLinesFromDir 从目录读取所有文本行
// filter表示后缀名
func readLinesFromDir(dir string, filter string) (lines []string, err error) {
	files, err := getFilesFromDir(dir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if !strings.HasSuffix(file, filter) {
			continue
		}
		linesTmp, err := readLinesFromFile(dir + "/" + file)
		if err != nil {
			return nil, err
		}
		lines = append(lines, linesTmp...)
	}
	return
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

// initUserList 初始化用户列表
func initUserList() (err error) {
	if settings.Get("user-check") == "on" {
		usersPath := settings.Get("config-dir") + "/users.list"
		err = importUsers(usersPath)
		if err != nil {
			return
		}
	}
	return err
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
	comm.Timeout = Timeout
	return nil
}

func execHosts() (err error) {
	hostsDir := settings.Get("config-dir") + "/hosts"
	count, err := readHosts(hostsDir)
	if err != nil {
		return err
	}
	logger.Info(count, " hosts lines imported")
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

// readLinesFromFile 从文件读取所有文本行
// 所有制表符会被替换为空格
// 首尾的空格会被去除
// # 会注释掉所有所在行剩下的内容
func readLinesFromFile(filePath string) ([]string, error) {
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

func readHosts(hostsDir string) (int, error) {
	hosts, err := readLinesFromDir(hostsDir, ".hosts")
	if err != nil {
		return 0, err
	}

	var count int
	for _, host := range hosts {
		err = handleSingleHost(host)
		if err != nil {
			return 0, err
		}
		count++
	}
	return count, nil
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

	if host == "" {
		return nil
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

// getFilesFromDir 获取指定目录中的所有文件
// 排除掉文件夹
func getFilesFromDir(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	var noDirFiles []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		noDirFiles = append(noDirFiles, file.Name())
	}
	return noDirFiles, nil
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

// CreateETArg 构建ET.Arg
func CreateETArg() *comm.Arg {
	users := comm.UsersArg{
		LocalUser:  LocalUser,
		ValidUsers: Users,
	}
	connArg := comm.ConnArg{
		RemoteIPE: settings.Get("relay"),
		Head:      settings.Get("head"),
	}
	smartArg := comm.SmartArg{
		ProxyStatus:   ProxyStatus,
		LocalLocation: settings.Get("location"),
	}

	return &comm.Arg{
		ConnArg:  connArg,
		SmartArg: smartArg,
		UsersArg: users,
		IPType:   settings.Get("ip-type"),
	}
}
