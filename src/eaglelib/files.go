package eaglelib

import (
	"os"
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
