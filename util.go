package oax

import (
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"sort"
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

type FileInfo struct {
	FileFullPath string
	FileName     string
}

func ListFiles(dir string) (fileInfos []FileInfo, err error) {
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			fileInfo := FileInfo{
				FileName:     filepath.Base(path),
				FileFullPath: path,
			}
			fileInfos = append(fileInfos, fileInfo)
		}

		return nil
	})

	if err != nil {
		return
	}

	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].FileName > fileInfos[j].FileName
	})

	return
}
