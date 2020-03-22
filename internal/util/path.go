package util

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//获取当前执行目录
func GetAppPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, string(os.PathSeparator))
	return path[:index]
}
