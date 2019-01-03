/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-02 12:42:49
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-03 23:25:57
 */

package cmd

import (
	"fmt"
	"os"
	"strconv"

	"../eagletunnel"
)

// Check check命令
func Check(args []string) {
	if len(args) < 3 {
		fmt.Println("no cmd for et-check")
		return
	}
	theType := eagletunnel.ParseEtCheckType(args[2])
	switch theType {
	case eagletunnel.EtCheckPING:
		ping()
	case eagletunnel.EtCheckAuth:
		auth()
	case eagletunnel.EtCheckVERSION:
		version()
	case eagletunnel.EtCheckSPEED:
		speed()
	default:
		fmt.Println("invalid check command")
	}
	os.Exit(0)
}

// ping 发送Ping请求并打印结果
func ping() {
	var time int
	var success int
	timeSig := make(chan string)
	for i := 0; i < 10; i++ {
		go eagletunnel.SendEtCheckPingReq(timeSig)
	}
	for i := 0; i < 10; i++ {
		timeStr := <-timeSig
		theTime, err := strconv.ParseInt(timeStr, 10, 32)
		if err != nil {
			fmt.Println("PING ", err.Error())
		} else {
			time += int(theTime)
			success++
			fmt.Println("PING ", theTime, "ms")
		}
	}
	fmt.Println("--- ping statistics ---")
	fmt.Println("10 messages transmitted, ", success, " received, ", (10-success)*10, "% loss")
}

func auth() {
	reply := eagletunnel.SendEtCheckAuthReq()
	fmt.Println(reply)
}

func version() {
	reply := eagletunnel.SendEtCheckVersionReq()
	fmt.Println(reply)
}

func speed() {
	reply := eagletunnel.SendEtCheckSpeedReq()
	speed, err := strconv.ParseUint(reply, 10, 64)
	if err != nil {
		fmt.Println(reply)
		return
	}
	if speed < 1024 {
		fmt.Println(speed, "Byte/s")
		return
	}
	speed /= 1024
	if speed < 1024 {
		fmt.Println(speed, "KB/s")
		return
	}
	speed /= 1024
	fmt.Println(speed, "MB/s")
	return
}
