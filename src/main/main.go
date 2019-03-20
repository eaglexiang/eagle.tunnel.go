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
	"github.com/eaglexiang/eagle.tunnel.go/src/logger"
	mycmd "github.com/eaglexiang/eagle.tunnel.go/src/mycmd"
	settings "github.com/eaglexiang/go-settings"
)

var service *etcore.Service

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("error: no arg for Eagle Tunnel")
		return
	}

	switch args[0] {
	case "check":
		err := Init(args[2:])
		if err != nil {
			logger.Error(err)
			return
		}
		mycmd.Check(args[1])
	default:
		err := Init(args)
		if err != nil {
			if err.Error() != "no need to continue" {
				logger.Error(err)
			}
			return
		}
		fmt.Println(settings.ToString())
		service = etcore.CreateService()
		defer service.Close()
		go core()
		checkSig()
	}
}

func checkSig() {
	fmt.Println("press Ctrl + C to quit")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	fmt.Println("stoping...")
}

// Init 初始化参数系统
func Init(args []string) error {
	err := mycmd.ImportArgs(args)
	if err != nil {
		return err
	}
	return etcore.ExecConfig()
}

func core() {
	err := service.Start()
	if err != nil {
		logger.Error(err)
	}
}
