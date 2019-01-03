/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-27 08:38:06
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-03 16:19:58
 */

package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"./cmd"
	"./eagletunnel"
)

var relayer = eagletunnel.Relayer{}

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("error: no arg for Eagle Tunnel")
	}

	switch args[1] {
	case "check":
		Init(args[2:])
		cmd.Check(args[:3])
	default:
		if Init(args) {
			fmt.Println(eagletunnel.SprintConfig())
			go relayer.Start()
		}
	}

	checkSig()
}

func checkSig() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	relayer.Close()
	time.Sleep(time.Second)
}

// Init 初始化参数系统
func Init(args []string) bool {
	err := cmd.ImportArgs(args)
	if err != nil {
		if err.Error() == "no need to continue" {
			return false
		}
		panic(err)
	}
	return true
}

func quit(relayer *eagletunnel.Relayer) {
	relayer.Close()
	time.Sleep(time.Second)
	os.Exit(0)
}
