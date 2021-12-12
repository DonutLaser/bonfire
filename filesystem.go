package main

import (
	"io/fs"
	"os"
	"path"
	"strconv"
	"strings"
)

type FileType int32

const (
	FileType_Default = iota
	FileType_Exe     = iota
	FileType_Image   = iota
)

func ReadFile(fullPath string) string {
	contents, err := os.ReadFile(fullPath)
	if err != nil {
		NotifyError(err.Error())
		return ""
	}

	return string(contents)
}

func WriteFile(fullPath string, contents string) {
	err := os.WriteFile(fullPath, []byte(contents), 0644)
	if err != nil {
		NotifyError(err.Error())
	}
}

func ReadDirectory(fullPath string) ([]fs.DirEntry, bool) {
	dir, err := os.Open(fullPath)
	if err != nil {
		NotifyError(err.Error())
		return []fs.DirEntry{}, false
	}
	defer dir.Close()

	items, err := dir.ReadDir(-1)
	if err != nil {
		NotifyError(err.Error())
		return []fs.DirEntry{}, false
	}

	return items, true
}

func GetDirectorySize(fullPath string) (result int64) {
	items, _ := ReadDirectory(fullPath)
	for _, item := range items {
		if item.IsDir() {
			result += GetDirectorySize(path.Join(fullPath, item.Name()))
		} else {
			info, err := item.Info()
			if err != nil {
				NotifyError(err.Error())
				continue
			}

			result += info.Size()
		}
	}

	return
}

func DuplicateFile(dirname string, filename string) (bool, string) {
	name := GetAvailableFileName(dirname, filename)

	data, err := os.ReadFile(path.Join(dirname, filename))
	if err != nil {
		NotifyError(err.Error())
		return false, ""
	}

	err = os.WriteFile(path.Join(dirname, name), data, 0644)
	if err != nil {
		NotifyError(err.Error())
		return false, ""
	}

	return true, name
}

func GetAvailableFileName(dirName string, filename string) (result string) {
	fullPath := path.Join(dirName, filename)

	extension := path.Ext(filename)
	name := strings.TrimSuffix(path.Base(filename), extension)

	fileDoesNotExist := true
	count := 0
	for fileDoesNotExist {
		_, err := os.Stat(fullPath)
		if err == nil {
			count += 1
			fullPath = path.Join(dirName, name+" ("+strconv.Itoa(count)+")"+extension)
		} else {
			fileDoesNotExist = false

			if count > 0 {
				result = name + " (" + strconv.Itoa(count) + ")" + extension
			} else {
				result = filename
			}
		}
	}

	return
}

func CreateNewFile(dirname string) (bool, string) {
	name := GetAvailableFileName(dirname, "New File")

	err := os.WriteFile(path.Join(dirname, name), []byte(""), 0644)
	if err != nil {
		NotifyError(err.Error())
		return false, ""
	}

	return true, name
}

func CreateNewFolder(dirname string, defaultName string) (bool, string) {
	name := GetAvailableFileName(dirname, defaultName)

	err := os.Mkdir(path.Join(dirname, name), 0755)
	if err != nil {
		NotifyError(err.Error())
		return false, ""
	}

	return true, name
}

func DoesFileExist(fullPath string) bool {
	_, err := os.Stat(fullPath)
	return err == nil
}

func GetFileType(fullPath string) ItemType {
	stat, err := os.Stat(fullPath)
	if err != nil {
		NotifyError(err.Error())
		return ItemTypeFile
	}

	if stat.IsDir() {
		return ItemTypeFolder
	}

	return ItemTypeFile
}

func GetAvailableDrives() (result []string) {
	for _, drive := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		file, err := os.Open(string(drive) + ":\\")
		if err == nil {
			result = append(result, string(drive))
			file.Close()
		}
	}

	return
}
