package paths

import (
	"path/filepath"

	"github.com/adrg/xdg"
)

const AppName = "aiodl"

func ConfigFile(name string) (string, error) {
	return xdg.ConfigFile(filepath.Join(AppName, name))
}

func CacheFile(name string) (string, error) {
	return xdg.CacheFile(filepath.Join(AppName, name))
}

func DataFile(name string) (string, error) {
	return xdg.DataFile(filepath.Join(AppName, name))
}

func ConfigDir() string {
	return filepath.Join(xdg.ConfigHome, AppName)
}

func CacheDir() string {
	return filepath.Join(xdg.CacheHome, AppName)
}

func DownloadDir() string {
	return filepath.Join(xdg.UserDirs.Download, AppName)
}
