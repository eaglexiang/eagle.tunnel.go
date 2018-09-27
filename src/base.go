package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// FileExsits 判断文件是否存在
func FileExsits(path string) bool {
	f, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return !f.IsDir()
		}
		return false
	}
	return !f.IsDir()
}

// GetCurrentPath 获取当前路径
func GetCurrentPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	i := strings.LastIndex(path, "/")
	if i < 0 {
		i = strings.LastIndex(path, "\\")
	}
	if i < 0 {
		return "", errors.New(`error: Can't find "/" or "\"`)
	}
	return string(path[0 : i+1]), nil
}
