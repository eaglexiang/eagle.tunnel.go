package config

import (
	"io/ioutil"
	"path"
	"path/filepath"
	"plugin"
	"strings"

	"github.com/eaglexiang/eagle.tunnel.go/src/core/protocols/et/comm"
	"github.com/eaglexiang/go-logger"
	"github.com/eaglexiang/go-settings"
)

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

func readConfigDir() (err error) {
	if !finishConfigDir() {
		return nil
	}
	if err = initUserList(); err != nil {
		return
	}
	if err = readClearDomains(); err != nil {
		return
	}
	// hosts文件
	if err = execHosts(); err != nil {
		return
	}
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
	proxyDomainsPath := path.Join(settings.Get("config-dir"), "/proxylists")
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

func execMods() (err error) {
	if modsDir := settings.Get("mod-dir"); modsDir != "" {
		err = ImportMods(modsDir)
		if err != nil {
			return err
		}
	}
	return nil
}

func execHosts() (err error) {
	hostsDir := path.Join(settings.Get("config-dir"), "/hosts")
	count, err := readHosts(hostsDir)
	if err != nil {
		return err
	}
	logger.Info(count, " hosts lines imported")
	return nil
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
