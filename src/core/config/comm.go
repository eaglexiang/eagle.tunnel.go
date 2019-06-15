package config

import (
	"bufio"
	"io/ioutil"
	"os"
	"strings"
)

// readLinesFromDir 从目录读取所有文本行
// filter表示后缀名
func readLinesFromDir(dir string, filter string) (lines []string) {
	files := getFilesFromDir(dir)
	for _, file := range files {
		if !strings.HasSuffix(file, filter) {
			continue
		}
		linesTmp := readLinesFromFile(dir + "/" + file)
		lines = append(lines, linesTmp...)
	}
	return
}

// readLinesFromFile 从文件读取所有文本行
// 所有制表符会被替换为空格
// 首尾的空格会被去除
// # 会注释掉所有所在行剩下的内容
func readLinesFromFile(filePath string) []string {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
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
	err = scanner.Err()
	if err != nil {
		panic(err)
	}
	return lines
}

// getFilesFromDir 获取指定目录中的所有文件
// 排除掉文件夹
func getFilesFromDir(dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	var noDirFiles []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		noDirFiles = append(noDirFiles, file.Name())
	}
	return noDirFiles
}
