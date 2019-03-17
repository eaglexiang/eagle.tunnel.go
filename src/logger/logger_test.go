/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-03-17 15:52:23
 * @LastEditTime: 2019-03-17 15:57:05
 */

package logger

import (
	"fmt"
	"testing"

	"github.com/eaglexiang/go-settings"
)

func Test_logger(t *testing.T) {
	settings.Set("debug", "off")
	print()
	settings.Set("debug", "error")
	print()
	settings.Set("debug", "warning")
	print()
	settings.Set("debug", "info")
	print()
	settings.Set("debug", "on")
	print()
}

func print() {
	fmt.Println("当前日志级别： " + settings.Get("debug"))
	Error("测试错误")
	Warning("测试警告")
	Info("测试消息")
}
