/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-27 09:42:11
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-02-21 15:19:22
 */

package cmd

import (
	"fmt"

	myversion "github.com/eaglexiang/go-version"
)

// ProgramVersion 程序版本
var ProgramVersion, _ = myversion.CreateVersion("0.9")

// PrintVersion 打印版本信息
func PrintVersion(programVersion, protocolVersion, ProtocolCompatibleVersion string) {
	fmt.Println(
		"eagle tunnel version:\t", programVersion, "\n",
		"protocol version:\t", protocolVersion, "\n",
		"protocol compatible version:\t", ProtocolCompatibleVersion,
	)
}
