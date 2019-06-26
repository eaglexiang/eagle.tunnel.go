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

func readConfigDir() {
	if !finishConfigDir() {
		return
	}

	initUserList()
	readClearDomains()
	// hosts文件
	importHosts()
	// 导入Mods
	execMods()
}

// readClearDomains 读取明确连接规则的域名列表
func readClearDomains() {
	readDirectlist()
	readProxylist()
}

// readWhitelist 读取强制代理的域名列表
func readProxylist() {
	proxyDomainsPath := path.Join(settings.Get("config-dir"), "/proxylists")
	proxyDomains := readLinesFromDir(proxyDomainsPath, ".txt")

	for _, domain := range proxyDomains {
		comm.ProxyDomains.ReverseGrow(domain)
	}

	logger.Info(comm.ProxyDomains.Count(), " proxy-domains imported")
}

// readDirectlist 读取强制直连的域名列表
func readDirectlist() {
	directDomainsPath := settings.Get("config-dir") + "/directlists"
	directDomains := readLinesFromDir(directDomainsPath, ".txt")

	for _, domain := range directDomains {
		comm.DirectDomains.ReverseGrow(domain)
	}

	logger.Info(comm.DirectDomains.Count(), " direct-domains imported")

}

// ImportMods 导入Mods
func ImportMods(modsDir string) {
	files, err := ioutil.ReadDir(modsDir)
	if err != nil {
		panic(err)
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
				panic(err)
			}
		}
	}
}

func execMods() {
	if modsDir := settings.Get("mod-dir"); modsDir != "" {
		ImportMods(modsDir)
	}
}

func importHosts() {
	hostsDir := path.Join(settings.Get("config-dir"), "/hosts")
	count := readHosts(hostsDir)
	logger.Info(count, " hosts lines imported")
}

func readHosts(hostsDir string) int {
	hosts := readLinesFromDir(hostsDir, ".hosts")

	var count int
	for _, host := range hosts {
		handleSingleHost(host)
		count++
	}
	return count
}
