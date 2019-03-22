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
	"github.com/eaglexiang/eagle.tunnel.go/src/core/protocols/et/cmd"
	et "github.com/eaglexiang/eagle.tunnel.go/src/core/protocols/et/core"
	myet "github.com/eaglexiang/eagle.tunnel.go/src/core/protocols/et/core"
	cipher "github.com/eaglexiang/go-cipher"
	settings "github.com/eaglexiang/go-settings"
)

// Check check命令
func Check(arg string) {
	theType := cmd.ParseEtCheckType(arg)
	switch theType {
	case cmd.EtCheckPING:
		ping()
	case cmd.EtCheckAUTH:
		auth()
	case cmd.EtCheckVERSION:
		version()
	case cmd.EtCheckUSERS:
		users()
	default:
		fmt.Println("invalid check command")
	}
	os.Exit(0)
}

// ping 发送Ping请求并打印结果
func ping() {
	createET()
	var time int
	var success int
	timeSig := make(chan string)
	for i := 0; i < 10; i++ {
		go cmd.SendEtCheckPingReq(timeSig)
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
	createET()
	reply := cmd.SendEtCheckAuthReq()
	fmt.Println(reply)
}

func version() {
	createET()
	reply, err := cmd.SendEtCheckVersionReq()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(reply)
	}
}

func users() {
	fmt.Println("USERS: ")
	fmt.Println("--- ---")

	createET()
	reply, err := cmd.SendEtCheckUsersReq()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(reply)
	}
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
	et := et.NewET(e)
	return et
}
