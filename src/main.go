/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-27 08:38:06
 * @LastEditors: EagleXiang
 * @LastEditTime: 2018-12-27 09:41:09
 */

package main

import (
	"fmt"
	"os"

	"./eaglelib/src"
	"./eagletunnel"
)

// ProgramVersion 程序版本
var ProgramVersion, _ = eaglelib.CreateVersion("0.7")

func main() {
	err := ImportArgs(os.Args)
	if err != nil {
		if err.Error() == "no need to continue" {
			return
		}
		panic(err)
	}

	relayer := eagletunnel.Relayer{}
	relayer.Start()
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
	err := eagletunnel.StartUI()
	if err != nil {
		fmt.Println(err)
	}
}
