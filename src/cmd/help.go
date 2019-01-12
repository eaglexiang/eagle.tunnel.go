/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-27 06:14:23
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-13 06:30:15
 */

package cmd

import (
	"fmt"
)

// PrintHelpMain 打印帮助信息主页
func PrintHelpMain() {
	text :=
		`Normal Usage: et [options...]
		
	-c	--config	path of config file
	-h	--help		print this info
	-l	--listen	ip(:port)
	-r	--relayer	ip(:port)
	-s	--proxy-status	smart/enable
	-u	--user		username:password
	-v	--version	print version of this program
		--config-dir	directory of extension config files
		--data-key	default:34
		--et		on/off
		--head		default:eagle_tunnel
		--http		on/off
		--socks		on/off
		--user-check	on/off

Command Usage: et [command...]

	check	run et check --help to get more information`
	fmt.Println(text)
}
