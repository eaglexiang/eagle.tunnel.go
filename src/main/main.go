/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-27 08:38:06
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-03-16 15:14:41
 */

package main

import (
	"fmt"
	"os"
	"os/signal"

	etcore "github.com/eaglexiang/eagle.tunnel.go/src/core/core"
	mycmd "github.com/eaglexiang/eagle.tunnel.go/src/mycmd"
	settings "github.com/eaglexiang/go-settings"
)

var service *etcore.Service

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("error: no arg for Eagle Tunnel")
	}

	switch args[1] {
	case "check":
		Init(args[2:])
		mycmd.Check(args[:3])
	default:
		err := Init(args)
		if err != nil {
			if err.Error() != "no need to continue" {
				fmt.Println(err)
			}
			return
		}
		fmt.Println(settings.ToString())
		service = etcore.CreateService()
		go core()
		fmt.Println("press Ctrl + C to quit")
		checkSig()
	}
}

func checkSig() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	fmt.Println("stoping...")
	service.Close()
}

// Init 初始化参数系统
func Init(args []string) error {
	err := mycmd.ImportArgs(args)
	return err
}

func core() {
	err := service.Start()
	if err != nil {
		fmt.Println("error: ", err)
	}
}
