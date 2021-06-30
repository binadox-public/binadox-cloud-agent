package engine

import (
	"github.com/adrg/xdg"
	"path"
)

var workDir string = path.Join(xdg.CacheHome, "binadox-agent")

func SetWorkDir(wd string) {
	workDir = wd
}
func GetWorkDir() string {
	return workDir
}

func GetCacheDir() string {
	return path.Join(workDir, "cache")
}

func GetUpdaterDir() string {
	return path.Join(workDir, "updates")
}

