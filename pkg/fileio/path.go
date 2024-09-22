package fileio

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/adrg/xdg"
)

var (
	execPath, _ = os.Executable()
	baseName    = trimSuffix(filepath.Base(execPath), ".exe", runtime.GOOS == "windows")
	dataDir     = filepath.Join(xdg.DataHome, baseName)
)

func SetDataDir(path string) {
	dataDir = path
}

func DataFile(name string) string {
	return filepath.Join(dataDir, name)
}

func trimSuffix(s, suffix string, fold ...bool) string {
	if len(s) >= len(suffix) {
		var hasSuffix bool
		if len(fold) > 0 && fold[0] {
			hasSuffix = strings.EqualFold(s[len(s)-len(suffix):], suffix)
		} else {
			hasSuffix = s[len(s)-len(suffix):] == suffix
		}
		if hasSuffix {
			return s[:len(s)-len(suffix)]
		}
	}
	return s
}
