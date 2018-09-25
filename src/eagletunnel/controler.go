package eagletunnel

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var ConfigPath string
var ConfigDir string
var ConfigKeyValues map[string]string

var EncryptKey byte
var EnableUserCheck bool

var EnableSOCKS5 bool
var EnableHTTP bool
var EnableET bool

var PROXY_STATUS int

func Init(filePath string) error {
	ConfigPath = filePath
	allConfLines, err := readLines(ConfigPath)
	if err != nil {
		fmt.Println("failed to read " + ConfigPath)
	}

	ConfigKeyValues, _ := getKeyValues(allConfLines)

	var ok bool

	ConfigDir, ok = ConfigKeyValues["config-dir"]
	if !ok {
		ConfigDir = filepath.Dir(ConfigPath)
		fmt.Println("default config-dir -> ", ConfigDir)
	}

	EnableUserCheck = false
	var enableUsercheck string
	enableUsercheck, ok = ConfigKeyValues["user-check"]
	if ok {
		EnableUserCheck = enableUsercheck == "on"
	}

	if EnableUserCheck {
		usersPath := ConfigDir + "/users.list"
		err = importUsers(usersPath)
		if err != nil {
			fmt.Println(err)
		}
	}

	var user string
	user, ok = ConfigKeyValues["user"]
	if ok {
		LocalUser, err = ParseEagleUser(user, nil)
		if err != nil {
			fmt.Println(err)
		}
	}

	go CheckSpeedOfUsers()

	EncryptKey = 0x22
	var encryptKey string
	encryptKey, ok = ConfigKeyValues["data-key"]
	if ok {
		var _encryptKey uint64
		_encryptKey, err = strconv.ParseUint(encryptKey, 16, 8)
		if err != nil {
			return err
		}
		EncryptKey = byte(uint8(_encryptKey))
	}

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

	PROXY_STATUS = ProxyENABLE
	var status string
	status, ok = ConfigKeyValues["proxy-status"]
	if ok {
		switch status {
		case "enable":
			PROXY_STATUS = ProxyENABLE
		case "smart":
			PROXY_STATUS = ProxySMART
		default:
			PROXY_STATUS = ProxyENABLE
		}
	}

	whiteDomainsPath := ConfigDir + "/whitelist_domain.txt"
	WhitelistDomains, _ = readLines(whiteDomainsPath)

	readHosts(ConfigDir)

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
		user, err = ParseEagleUser(line, nil)
		if err != nil {
			return err
		}
		Users[user.ID] = user
	}
	return err
}

func getKeyValues(lines []string) (map[string]string, []string) {
	keyValues := make(map[string]string)
	keys := make([]string, 0)
	for _, line := range lines {
		keyValue := strings.Split(line, "=")
		if len(keyValue) >= 2 {
			value := keyValue[1]
			for _, item := range keyValue[2:] {
				value += "=" + item
			}
			key := strings.TrimSpace(keyValue[0])
			keys = append(keys, key)
			value = strings.TrimSpace(value)
			keyValues[key] = value
		}
	}
	return keyValues, keys
}

func exportKeyValues(keyValues *map[string]string, keys []string) string {
	var result string
	for _, key := range keys {
		result += key + " = " + (*keyValues)[key] + "\r\n"
	}
	return result
}

func SetRelayer(remoteIpe string) {
	items := strings.Split(remoteIpe, ":")
	RemoteAddr = strings.TrimSpace(items[0])
	if len(items) >= 2 {
		RemotePort = strings.TrimSpace(items[1])
	} else {
		RemotePort = "8080"
	}
}

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

func readHosts(configDir string) {
	hostsDir := configDir + "/hosts"

	hostsFiles := getHostsList(hostsDir)

	var hosts []string
	for _, file := range hostsFiles {
		newHosts, err := readLines(hostsDir + "/" + file)
		if err == nil {
			hosts = append(hosts, newHosts...)
		}
	}

	for _, host := range hosts {
		items := strings.Split(host, " ")
		if len(items) >= 2 {
			domain := strings.TrimSpace(items[0])
			ip := strings.TrimSpace(items[1])
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
	} else {
		var hosts []string
		for _, file := range files {
			if !file.IsDir() {
				filename := file.Name()
				hosts = append(hosts, filename)
			}
		}
		return hosts
	}
}
