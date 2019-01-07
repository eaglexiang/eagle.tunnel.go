/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-27 09:42:11
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-07 16:31:54
 */

package cmd

import (
	"fmt"

	"../eaglelib/src"
)

// ProgramVersion 程序版本
var ProgramVersion, _ = eaglelib.CreateVersion("0.7.2")

// PrintVersion 打印版本信息
func PrintVersion(programVersion, protocolVersion, ProtocolCompatibleVersion string) {
	fmt.Println(
		"eagle tunnel version:\t", programVersion, "\n",
		"protocol version:\t", protocolVersion, "\n",
		"protocol compatible version:\t", ProtocolCompatibleVersion,
	)
}
