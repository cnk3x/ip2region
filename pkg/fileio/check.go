package fileio

import (
	"fmt"
	"os"
)

func CheckExist(filePath string, onNonExist func() error) (err error) {
	stat, e := os.Stat(filePath)

	if e != nil {
		if os.IsNotExist(e) {
			err = onNonExist()
		} else {
			err = fmt.Errorf("无法访问地址库文件: %w", e)
		}
		return
	}

	if !stat.Mode().IsRegular() {
		err = fmt.Errorf("错误的地址库文件路径: %s", filePath)
	}
	return
}
