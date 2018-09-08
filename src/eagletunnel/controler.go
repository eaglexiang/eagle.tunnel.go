package eagletunnel

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var ConfigPath string
var ConfigDir string
var ConfigKeyValues map[string]string

var EncryptKey byte
var Users map[string]*EagleUser
var EnableUserCheck bool

var EnableSOCKS5 bool
var EnableHTTP bool
var EnableET bool

var PROXY_STATUS int

func Init(filePath string) error {
	ConfigPath = filePath
	allConfLines, err := readLines(ConfigPath)
	if err != nil {
		return errors.New("failed to read " + ConfigPath)
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
			return err
		}
	}

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

	var user string
	user, ok = ConfigKeyValues["user"]
	if ok {
		LocalUser, err = ParseEagleUser(user)
		if err != nil {
			return err
		}
	}

	var localIpe string
	localIpe, ok = ConfigKeyValues["listen"]
	if ok {
		items := strings.Split(localIpe, ":")
		LocalAddr = items[0]
		if len(items) >= 2 {
			LocalPort = items[1]
		} else {
			LocalPort = "8080"
		}
	} else {
		LocalAddr = "0.0.0.0"
		LocalPort = "8080"
	}

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
			items := strings.Split(remoteIpe, ":")
			RemoteAddr = items[0]
			if len(items) >= 2 {
				RemotePort = items[1]
			} else {
				RemotePort = "8080"
			}
		}
	}

	PROXY_STATUS = PROXY_ENABLE
	var status string
	status, ok = ConfigKeyValues["proxy-status"]
	if ok {
		switch status {
		case "enable":
			PROXY_STATUS = PROXY_ENABLE
		case "smart":
			PROXY_STATUS = PROXY_SMART
		default:
			PROXY_STATUS = PROXY_ENABLE
		}
	}

	whiteDomainsPath := ConfigDir + "/whitelist_domain.txt"
	WhitelistDomains, _ = readLines(whiteDomainsPath)

	return err
}

func SPrintConfig() string {
	var status string
	switch PROXY_STATUS {
	case PROXY_SMART:
		status = "smart"
	case PROXY_ENABLE:
		status = "enable"
	default:
	}

	var localId string
	if LocalUser != nil {
		localId = LocalUser.Id
	}

	var configStr string
	configStr += "ConfigPath: " + ConfigPath + "\n"
	configStr += "ConfigDir: " + ConfigDir + "\n"
	configStr += "RemoteAddr: " + RemoteAddr + ":" + RemotePort + "\n"
	configStr += "LocalAddr: " + LocalAddr + ":" + LocalPort + "\n"
	configStr += "EncryptKey: " + strconv.Itoa(int(EncryptKey)) + "\n"
	configStr += "Count of Users: " + strconv.Itoa(len(Users)) + "\n"
	configStr += "Local User: " + localId + "\n"
	configStr += "HTTP: " + strconv.FormatBool(EnableHTTP) + "\n"
	configStr += "SOCKS5: " + strconv.FormatBool(EnableSOCKS5) + "\n"
	configStr += "ET: " + strconv.FormatBool(EnableET) + "\n"
	configStr += "Status: " + status + "\n"
	return configStr
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
		} else {
			Users[user.Id] = user
		}
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
