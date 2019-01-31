/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-27 08:38:06
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-31 21:13:01
 */

package main

import (
	"fmt"
	"os"
	"os/signal"

	"./cmd"
	"./etcore"
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
		cmd.Check(args[:3])
	default:
		err := Init(args)
		if err != nil {
			if err.Error() != "no need to continue" {
				fmt.Println(err)
			}
			return
		}
		fmt.Println(etcore.SprintConfig())
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
	err := cmd.ImportArgs(args)
	return err
}

func core() {
	err := service.Start()
	if err != nil {
		fmt.Println("error: ", err)
	}
}
