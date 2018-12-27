/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-27 06:14:23
 * @LastEditors: EagleXiang
 * @LastEditTime: 2018-12-27 09:42:18
 */

package shell

import (
	"fmt"
)

// PrintHelpMain 打印帮助信息主页
func PrintHelpMain() {
	text :=
		`Usage: et [options...] <config file>
		-h	--help
		-v	--version
		-l	--listen		127.0.0.1:8080
		-r	--relayer		127.0.0.1:8080
		-s	--proxy-status	smart/enable
		-u	--user			username:password
			--user-check	on/off
			--http			on/off
			--socks			on/off
			--et			on/off
			--data-key		34
			--head			eagle_tunnel
		-c	--config		/etc/eagle-tunnel.d/client.conf
			--config-dir	/etc/eagle-tunnel.d`
	fmt.Println(text)
}

func printHelpAsk() {
	fmt.Println(
		"Usage: et ask [options] <config file>\n",
		"\tauth\tcheck you local user\n",
		"\tping\tcheck your ping to remote relayer",
	)
}
