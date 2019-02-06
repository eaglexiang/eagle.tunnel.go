/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-27 05:42:47
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-02-07 00:01:18
 */

package cmd

import (
	"../etcore"
	"../etcore/et"
	"github.com/eaglexiang/go-settings"

	"errors"
	"strings"
)

// ImportArgs 解析并导入参数
func ImportArgs(argStrs []string) error {
	for i := 1; i < len(argStrs); i++ {
		if i%2 == 0 {
			// 奇数为参数名，偶数为参数值
			continue
		}

		argStrs[i] = toLongArg(argStrs[i])
		switch argStrs[i] {
		case "--help":
			PrintHelpMain()
			return errors.New("no need to continue")
		case "--version":
			PrintVersion(
				ProgramVersion.Raw,
				et.ProtocolVersion.Raw,
				et.ProtocolCompatibleVersion.Raw)
			return errors.New("no need to continue")
		default:
			err := importArg(argStrs, i)
			if err != nil {
				return err
			}
		}
	}

	err := etcore.ExecConfig()
	if err != nil {
		return err
	}

	return err
}

func toLongArg(shortArg string) string {
	switch shortArg {
	case "-h":
		return "--help"
	case "-v":
		return "--version"
	case "-l":
		return "--listen"
	case "-r":
		return "--relayer"
	case "-s":
		return "--proxy-status"
	case "-u":
		return "--user"
	case "-c":
		return "--config"
	default:
		return shortArg
	}
}

func toArgName(argStr string) (string, error) {
	if !strings.HasPrefix(argStr, "--") {
		return "", errors.New("toArgName -> invalid argStr: " + argStr)
	}
	return strings.TrimPrefix(argStr, "--"), nil
}

func importArg(argStrs []string, indexOfArg int) (err error) {
	key := argStrs[indexOfArg]
	key, err = toArgName(key)
	if err != nil {
		return errors.New("importArg -> " + err.Error())
	}
	indexOfValue := indexOfArg + 1
	if indexOfValue == len(argStrs) {
		return errors.New("importArg -> no value for arg: " + key)
	}
	value := argStrs[indexOfValue]
	settings.Set(key, value)
	return nil
}
