package main

import (
	"io/fs"
	"os"
	"path"
	"strconv"
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

func ReadDirectory(fullPath string) []fs.DirEntry {
	dir, err := os.Open(fullPath)
	if err != nil {
		NotifyError(err.Error())
		return []fs.DirEntry{}
	}
	defer dir.Close()

	items, err := dir.ReadDir(-1)
	if err != nil {
		NotifyError(err.Error())
		return []fs.DirEntry{}
	}

	return items
}

// @TODO (!important) This lags like hell if the directory is big
func GetDirectorySize(fullPath string) (result int64) {
	items := ReadDirectory(fullPath)
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
	// @TODO (!important) this should not add the numbers after the extension making the extension unusable

	fullPath := path.Join(dirName, filename)

	fileDoesNotExist := true
	count := 0
	for fileDoesNotExist {
		_, err := os.Stat(fullPath)
		if err == nil {
			count += 1
			fullPath = path.Join(dirName, filename+" ("+strconv.Itoa(count)+")")
		} else {
			fileDoesNotExist = false

			if count > 0 {
				result = filename + " (" + strconv.Itoa(count) + ")"
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

func CreateNewFolder(dirname string) (bool, string) {
	name := GetAvailableFileName(dirname, "New Folder")

	err := os.Mkdir(path.Join(dirname, name), 0755)
	if err != nil {
		NotifyError(err.Error())
		return false, ""
	}

	return true, name

}
