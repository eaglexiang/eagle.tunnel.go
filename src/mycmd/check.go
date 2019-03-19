/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-02 12:42:49
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-03-17 15:07:11
 */

package cmd

import (
	"fmt"
	"os"
	"strconv"

	etcore "github.com/eaglexiang/eagle.tunnel.go/src/core/core"
	myet "github.com/eaglexiang/eagle.tunnel.go/src/core/protocols/et"
	cipher "github.com/eaglexiang/go-cipher"
	settings "github.com/eaglexiang/go-settings"
)

// Check check命令
func Check(args []string) {
	if len(args) < 3 {
		fmt.Println("no cmd for et-check")
		return
	}
	theType := myet.ParseEtCheckType(args[2])
	switch theType {
	case myet.EtCheckPING:
		ping()
	case myet.EtCheckAUTH:
		auth()
	case myet.EtCheckVERSION:
		version()
	case myet.EtCheckUSERS:
		users()
	default:
		fmt.Println("invalid check command")
	}
	os.Exit(0)
}

// ping 发送Ping请求并打印结果
func ping() {
	et := createET()
	var time int
	var success int
	timeSig := make(chan string)
	for i := 0; i < 10; i++ {
		go myet.SendEtCheckPingReq(et, timeSig)
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
	et := createET()
	reply := myet.SendEtCheckAuthReq(et)
	fmt.Println(reply)
}

func version() {
	et := createET()
	reply := myet.SendEtCheckVersionReq(et)
	fmt.Println(reply)
}

func users() {
	fmt.Println("USERS: ")
	fmt.Println("--- ---")

	et := createET()
	reply := myet.SendEtCheckUsersReq(et)
	fmt.Println(reply)
}

func createET() *myet.ET {
	cipher.DefaultCipher = func() cipher.Cipher {
		cipherType := cipher.ParseCipherType(settings.Get("cipher"))
		switch cipherType {
		case cipher.SimpleCipherType:
			c := cipher.SimpleCipher{}
			c.SetKey(settings.Get("data-key"))
			return &c
		default:
			return nil
		}
	}

	e := etcore.CreateETArg()
	et := myet.CreateET(e)
	return et
}
