package main

import (
	"fmt"
	"os"

	"./eagletunnel"
)

func main() {
	args := os.Args
	var firstArg string
	if len(args) < 2 {
		firstArg = "./eagle-tunnel.conf"
	} else {
		firstArg = args[1]
	}
	switch firstArg {
	case "-u", "--ui":
		startUI()
	default:
		startTunnel(firstArg)
	}
}

func startUI() {
	_ = eagletunnel.Init("./eagle-tunnel.conf")
	err := eagletunnel.StartUI()
	if err != nil {
		fmt.Println(err)
	}
}

func startTunnel(pathOfConfig string) {
	err := eagletunnel.Init(pathOfConfig)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(eagletunnel.SPrintConfig())
	if err != nil {
		fmt.Println(err)
		return
	}
	relayer := eagletunnel.Relayer{}
	relayer.Start()
}
