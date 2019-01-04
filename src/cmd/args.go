/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-27 05:42:47
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-04 18:34:22
 */

package cmd

import (
	"../eagletunnel"

	"errors"
	"strings"
)

// ImportArgs 解析并导入参数
func ImportArgs(argStrs []string) error {
	indexOfConfig := findConfig(argStrs)

	var skip bool
	for i, v := range argStrs {
		if skip {
			skip = false
			continue
		}
		if i == 0 {
			continue
		}

		v := toLongArg(v)
		argStrs[i] = v
		switch v {
		case "--config":
			skip = true
			continue
		case "--help":
			PrintHelpMain()
			return errors.New("no need to continue")
		case "--version":
			PrintVersion(
				ProgramVersion.Raw,
				eagletunnel.ProtocolVersion.Raw,
				eagletunnel.ProtocolCompatibleVersion.Raw)
			return errors.New("no need to continue")
		default:
			var err error
			skip, err = importArg(argStrs, i)
			if err != nil {
				return err
			}
		}
	}

	if indexOfConfig > 0 {
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

	err := eagletunnel.ExecConfig()
	if err != nil {
		return err
	}

	return err
}

// importArg skip表示下个参数是否跳过
func importArg(argStrs []string, indexOfArg int) (skip bool, err error) {
	switch argStrs[indexOfArg] {
	case "--listen",
		"--relayer",
		"--proxy-status",
		"--user",
		"--http",
		"--socks",
		"--et",
		"--data-key",
		"--head",
		"--config-dir",
		"--user-check",
		"--speed-check",
		"--debug":
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
	argName := argStrs[indexOfArg]
	argName = strings.TrimPrefix(argName, "--")
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
