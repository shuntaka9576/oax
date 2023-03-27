package oax

import (
	"os"
	"os/user"
	"strings"
)

func createDirIfNotExist(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.Mkdir(dirPath, 0755)
	}
	return nil
}

func replaceTildeWithHomedir(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		homeDir := usr.HomeDir
		path = strings.Replace(path, "~", homeDir, 1)
	}

	return path, nil
}

func replaceHomedirWithTilde(path string) (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	homedir := usr.HomeDir

	if strings.HasPrefix(path, homedir) {
		return strings.Replace(path, homedir, "~", 1), nil
	}

	return path, nil
}
