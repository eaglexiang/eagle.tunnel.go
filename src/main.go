package main

import (
	"fmt"
	"os"

	"./eagletunnel"
)

const defaultPathOfConfig = "./config/client.conf"

func main() {
	args := os.Args
	var firstArg string
	if len(args) < 2 {
		firstArg = defaultPathOfConfig
	} else {
		firstArg = args[1]
	}

	switch firstArg {
	case "-u", "--ui":
		var secondArg string
		if len(args) <= 3 {
			secondArg = defaultPathOfConfig
		} else {
			secondArg = args[2]
		}
		startUI(secondArg)
	default:
		startTunnel(firstArg)
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
