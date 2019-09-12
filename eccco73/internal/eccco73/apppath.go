package eccco73

import (
	"os"
	"github.com/lxn/win"
	"syscall"
	"path/filepath"
	"os/user"
)

const (
	AppName = "eccco73"
)


func AppDataFileNameEnsureDir(filename string) string {
	return filepath.Join(EnsureAppDataDir(), filename)
}

func EnsureAppDataDir() string {
	var appDataDir string
	if appDataDir = os.Getenv("MYAPPDATA"); len(appDataDir) == 0 {
		var buf [win.MAX_PATH]uint16
		if !win.SHGetSpecialFolderPath(0, &buf[0], win.CSIDL_APPDATA, false) {
			panic("SHGetSpecialFolderPath failed")
		}
		appDataDir = syscall.UTF16ToString(buf[0:])
	}
	return ensureDir(filepath.Join(appDataDir, "Аналитприбор", AppName))
}

func EnsureAppDir() string {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	return ensureDir(filepath.Join(usr.HomeDir, "."+AppName))
}

func AppFileNameEnsureDir(filename string) string {
	return filepath.Join(EnsureAppDir(), filename)
}

func ensureDir(dir string) string {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) { // создать каталог если его нет
			os.Mkdir(dir, os.ModePerm)
		} else {
			panic(err)
		}
	}
	return dir
}


