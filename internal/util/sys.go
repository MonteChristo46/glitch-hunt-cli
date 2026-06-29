package util

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

func IsAdmin() bool {
	if runtime.GOOS == "windows" {
		_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
		return err == nil
	}
	u, err := user.Current()
	if err != nil {
		return false
	}
	return u.Uid == "0"
}

func GetExecutableDir() string {
	ex, err := os.Executable()
	if err != nil {
		return "."
	}
	realPath, err := filepath.EvalSymlinks(ex)
	if err != nil {
		return filepath.Dir(ex)
	}
	return filepath.Dir(realPath)
}

func GetRealUserHome() string {
	sudoUser := os.Getenv("SUDO_USER")
	if sudoUser != "" {
		u, err := user.Lookup(sudoUser)
		if err == nil {
			return u.HomeDir
		}
	}
	u, err := user.Current()
	if err == nil {
		return u.HomeDir
	}
	home, _ := os.UserHomeDir()
	return home
}
