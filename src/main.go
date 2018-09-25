package main

import (
	"errors"
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
	case "ask":
		err := ask(args)
		if err != nil {
			fmt.Println(err)
		}
	case "-u", "--ui":
		var secondArg string
		if len(args) < 3 {
			secondArg = defaultPathOfConfig
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
			pathOfConfig = defaultPathOfConfig
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
