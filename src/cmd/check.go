/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-02 12:42:49
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-02 14:48:32
 */

package cmd

import (
	"fmt"
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
	}
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
