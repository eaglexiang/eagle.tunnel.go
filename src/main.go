package main

import (
	"fmt"
	"os"

	"./eaglelib/src"
	"./eagletunnel"
)

// version 程序版本
var version, _ = eaglelib.CreateVersion("0.7")

func main() {
	args := os.Args
	var firstArg string
	if len(args) < 2 {
		firstArg = "-h"
	} else {
		firstArg = args[1]
	}

	if firstArg == "client" {
		firstArg = eagletunnel.DefaultClientConfig()
	} else if firstArg == "server" {
		firstArg = eagletunnel.DefaultServerConfig()
	}

	switch firstArg {
	case "ask":
		ask(args[2:]) // 过滤掉 'et ask'
	case "-h", "--help":
		printHelp(args)
	case "-v", "--version":
		printVersion()
	case "-u", "--ui":
		var secondArg string
		if len(args) < 3 {
			secondArg = eagletunnel.DefaultServerConfig()
		} else {
			secondArg = args[2]
		}
		startUI(secondArg)
	default:
		startTunnel(firstArg)
	}
}

func ask(args []string) {
	et := eagletunnel.EagleTunnel{}
	e := eagletunnel.NetArg{}
	e.TheType = eagletunnel.EtASK
	e.Args = args
	et.Send(&e)
	fmt.Println(e.Reply)
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
		printHelpMain()
		return
	}
	switch args[2] {
	case "ask":
		printHelpAsk()
	default:
		printHelpMain()
	}
}

func printHelpMain() {
	fmt.Println(
		"Usage: et [options...] <config file>\n",
		"\t-h,\t--help\tthis help text\n",
		"\t-v,\t--version\tprint version\n",
		"\tclient\tuse default client config file --> "+eagletunnel.DefaultClientConfig()+"\n",
		"\tserver\tuse default server config file --> "+eagletunnel.DefaultServerConfig()+"\n",
		"\task\tplease run \"et -h ask\" or \"et --help ask\"",
	)
}

func printHelpAsk() {
	fmt.Println(
		"Usage: et ask [options] <config file>\n",
		"\tauth\tcheck you local user\n",
		"\tping\tcheck your ping to remote relayer",
	)
}

func printVersion() {
	fmt.Println(
		"et version:\t", version.Raw, "\n",
		"protocol version:\t", eagletunnel.ProtocolVersion.Raw, "\n",
		"protocol compatible version:\t", eagletunnel.ProtocolCompatibleVersion.Raw,
	)
}
