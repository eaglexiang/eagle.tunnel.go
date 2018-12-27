/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-27 05:42:47
 * @LastEditors: EagleXiang
 * @LastEditTime: 2018-12-27 08:39:08
 */

package main

import (
	"./eagletunnel"
	"./shell"

	"errors"
	"fmt"
	"strings"
)

// ImportArgs 解析并导入参数
func ImportArgs(argStrs []string) error {
	indexOfConfig := findConfig(argStrs)
	if indexOfConfig >= 0 {
		err := checkIndex(argStrs, indexOfConfig)
		if err != nil {
			return err
		}
		err = eagletunnel.InitConfig(argStrs[indexOfConfig+1])
		if err != nil {
			return err
		}
	} else {
		eagletunnel.InitConfig("")
	}

	var skip bool
	for i, v := range argStrs {
		if skip {
			skip = false
			continue
		}
		if i == 0 {
			continue
		}
		if v == "-c" || v == "--config" {
			skip = true
			continue
		}
		var err error
		skip, err = importArg(argStrs, i)
		if err != nil {
			return err
		}
	}
	err := eagletunnel.ExcConfig()
	if err != nil {
		return err
	}

	fmt.Println(eagletunnel.SprintConfig())
	return err
}

// importArg skip表示下个参数是否跳过
func importArg(argStrs []string, indexOfArg int) (skip bool, err error) {
	switch argStrs[indexOfArg] {
	case "-h", "--help":
		shell.PrintHelpMain()
		return false, nil
	case "-v", "--version":
		shell.PrintVersion(ProgramVersion.Raw,
			eagletunnel.ProtocolVersion.Raw,
			eagletunnel.ProtocolCompatibleVersion.Raw)
		return false, nil
	case "-l", "--listen":
		return true, setKeyValue(argStrs, indexOfArg)
	case "-r", "--relayer":
		return true, setKeyValue(argStrs, indexOfArg)
	case "-s", "--proxy-status":
		return true, setKeyValue(argStrs, indexOfArg)
	case "-u", "--user":
		return true, setKeyValue(argStrs, indexOfArg)
	case "--http":
		return true, setKeyValue(argStrs, indexOfArg)
	case "--socks":
		return true, setKeyValue(argStrs, indexOfArg)
	case "--et":
		return true, setKeyValue(argStrs, indexOfArg)
	case "--data-key":
		return true, setKeyValue(argStrs, indexOfArg)
	case "--head":
		return true, setKeyValue(argStrs, indexOfArg)
	case "--config-dir":
		return true, setKeyValue(argStrs, indexOfArg)
	default:
		return false, errors.New("invalid arg: " + argStrs[indexOfArg])
	}
}

func checkIndex(argStrs []string, indexOfArg int) error {
	if len(argStrs) == indexOfArg+1 {
		return errors.New("no value for arg: " + argStrs[indexOfArg])
	}
	return nil
}

func setKeyValue(argStrs []string, indexOfArg int) error {
	arg := argStrs[indexOfArg]
	argName := strings.TrimPrefix(arg, "-")
	err := checkIndex(argStrs, indexOfArg)
	if err != nil {
		return err
	}
	eagletunnel.ConfigKeyValues[argName] = argStrs[indexOfArg+1]
	return nil
}

func findConfig(argStrs []string) int {
	for i, v := range argStrs {
		if v == "-c" || v == "--config" {
			return i
		}
	}
	return -1
}
