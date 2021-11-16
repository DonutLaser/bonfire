package main

import (
	"os"
)

func ReadFile(fullPath string) string {
	contents, err := os.ReadFile(fullPath)
	if err != nil {
		NotifyError(err.Error())
		return ""
	}

	return string(contents)
}
