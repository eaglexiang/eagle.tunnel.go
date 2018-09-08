package main

import (
	"fmt"
	"os"

	"./eagletunnel"
)

func main() {
	args := os.Args
	var configPath string
	if len(args) < 2 {
		configPath = "./eagle-tunnel.conf"
	} else {
		configPath = args[1]
	}
	err := eagletunnel.Init(configPath)
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

	// err = eagletunnel.StartUI()
	// if err != nil {
	// 	fmt.Println(err)
	// }
}
