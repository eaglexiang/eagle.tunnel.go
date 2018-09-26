package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"./eagletunnel"
)

var defaultPathsOfClientConfig = []string{
	"./config/client.conf",
	"/etc/eagle-tunnel.d/client.conf"}
var defaultPathsOfServerConfig = []string{
	"./config/server.conf",
	"/etc/eagle-tunnel.d/server.conf"}

func main() {
	args := os.Args
	var firstArg string
	if len(args) < 2 {
		firstArg = "-h"
	} else {
		firstArg = args[1]
	}

	if firstArg == "client" {
		firstArg = defaultClientConfig()
	} else if firstArg == "server" {
		firstArg = defaultServerConfig()
	}

	switch firstArg {
	case "ask":
		err := ask(args)
		if err != nil {
			fmt.Println(err)
		}
	case "-h", "--help":
		printHelp(args)
	case "-u", "--ui":
		var secondArg string
		if len(args) < 3 {
			secondArg = defaultServerConfig()
		} else {
			secondArg = args[2]
		}
		startUI(secondArg)
	default:
		startTunnel(firstArg)
	}
}

func ask(args []string) error {
	if len(args) < 3 {
		return errors.New("no arg for et ask")
	}
	askType := eagletunnel.ParseEtAskType(args[2])
	switch askType {
	case eagletunnel.EtAskAUTH:
		var pathOfConfig string
		if len(args) < 4 {
			pathOfConfig = defaultClientConfig()
		} else {
			pathOfConfig = args[3]
		}
		err := eagletunnel.Init(pathOfConfig)
		if err != nil {
			return err
		}
		et := eagletunnel.EagleTunnel{}
		e := eagletunnel.NetArg{TheType: []int{eagletunnel.EtASK, eagletunnel.EtAskAUTH}}
		_ = et.Send(&e)
		return errors.New(e.Reply)
	default:
		return nil
	}
}

func startUI(pathOfConfig string) {
	_ = eagletunnel.Init(pathOfConfig)
	err := eagletunnel.StartUI()
	if err != nil {
		fmt.Println(err)
	}
}

func startTunnel(pathOfConfig string) {
	err := eagletunnel.Init(pathOfConfig)
	fmt.Println(eagletunnel.SPrintConfig())
	if err != nil {
		fmt.Println(err)
		return
	}
	relayer := eagletunnel.Relayer{}
	relayer.Start()
}

func printHelp(args []string) {
	if len(args) < 3 {
		fmt.Println(
			"Usage: et [options...] <config file>\n" +
				"\t-h,\t--help\tThis help text\n" +
				"\tclient\tuse default client config file --> " + defaultClientConfig() + "\n" +
				"\tserver\tuse default server config file --> " + defaultServerConfig() + "\n" +
				"\task\tplease run \"et -h ask\" or \"et --help ask\"\n")
		return
	}
	switch args[2] {
	case "ask":
		printHelpAsk()
	}
}

func printHelpAsk() {
	fmt.Println(
		"Usage: et ask [options] <config file>\n" +
			"\tauth\tcheck you local user")
}

func defaultClientConfig() string {
	for _, path := range defaultPathsOfClientConfig {
		if fileExsits(path) {
			return path
		}
	}
	path, err := GetCurrentPath()
	if err != nil {
		return ""
	}
	return path
}

func defaultServerConfig() string {
	for _, path := range defaultPathsOfServerConfig {
		if fileExsits(path) {
			return path
		}
	}
	path, err := GetCurrentPath()
	if err != nil {
		return ""
	}
	return path
}

func fileExsits(path string) bool {
	f, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return !f.IsDir()
		}
		return false
	}
	return !f.IsDir()
}

func GetCurrentPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	i := strings.LastIndex(path, "/")
	if i < 0 {
		i = strings.LastIndex(path, "\\")
	}
	if i < 0 {
		return "", errors.New(`error: Can't find "/" or "\".`)
	}
	return string(path[0 : i+1]), nil
}
