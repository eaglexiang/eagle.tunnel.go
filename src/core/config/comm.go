package config

import (
	"bufio"
	"io/ioutil"
	"os"
	"strings"

	"github.com/eaglexiang/go-logger"
)

// readLinesFromDir 从目录读取所有文本行
// filter表示后缀名
func readLinesFromDir(dir string, filter string) (lines []string, err error) {
	files, err := getFilesFromDir(dir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if !strings.HasSuffix(file, filter) {
			continue
		}
		linesTmp, err := readLinesFromFile(dir + "/" + file)
		if err != nil {
			return nil, err
		}
		lines = append(lines, linesTmp...)
	}
	return
}

// readLinesFromFile 从文件读取所有文本行
// 所有制表符会被替换为空格
// 首尾的空格会被去除
// # 会注释掉所有所在行剩下的内容
func readLinesFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		items := strings.Split(line, "#")
		line = strings.TrimSpace(items[0])
		if line != "" {
			line = strings.Replace(line, "\t", " ", -1)
			lines = append(lines, line)
		}
	}
	return lines, scanner.Err()
}

// getFilesFromDir 获取指定目录中的所有文件
// 排除掉文件夹
func getFilesFromDir(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	var noDirFiles []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		noDirFiles = append(noDirFiles, file.Name())
	}
	return noDirFiles, nil
}
